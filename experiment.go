// Copyright 2013 SoundCloud, Rany Keddo. All rights reserved.  Use of this
// source code is governed by a license that can be found in the LICENSE file.

package bandit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"
)

// NewExperiment loads experiment `name` from the experiments source.
func NewExperiment(o Opener, name string) (*Experiment, error) {
	es, err := NewExperiments(o)
	if err != nil {
		return &Experiment{}, err
	}

	e, ok := (*es)[name]
	if !ok {
		return &Experiment{}, fmt.Errorf("could not find '%s' experiment", name)
	}

	return e, nil
}

// Experiment is a single experiment. Variations are in ascending ordinal
// sorting, where ordinals are contiguous and start at 1.
type Experiment struct {
	Name             string
	Strategy         Strategy
	Variations       Variations
	PreferredOrdinal int
}

// Select calls SelectArm on the strategy and returns the associated variation
func (e *Experiment) Select() Variation {
	selected := e.Strategy.SelectArm()
	if selected > len(e.Variations) {
		panic("selected impossible arm")
	}

	v, _ := e.GetVariation(selected)
	return v
}

// SelectTimestamped selects the appropriate variation given it's
// timestampedTag. A timestamped tag is a string in the form
// <tag>:<timestamp>. If the duration between <timestamp> and the current time
// is smaller than `d`, the given tagged is used to return variation. If it is
// larger, Select() is called instead.  If the `timestampedTag` argument is
// the blank string, Select() is called instead.
func (e *Experiment) SelectTimestamped(
	timestampedTag string,
	ttl time.Duration) (Variation, string, error) {
	now := time.Now().Unix()

	if timestampedTag == "" {
		selected := e.Select()
		return selected, makeTimestampedTag(selected, now), nil
	}

	tag, ts, err := TimestampedTagToTag(timestampedTag)
	if err != nil {
		return Variation{}, "", fmt.Errorf("bad timestamped tag: %s", err.Error())
	}

	// return the given timestamped tag
	if ttl > time.Since(time.Unix(ts, 0)) {
		v, err := e.GetTaggedVariation(tag)

		// could not get tagged variation. this can occurr when switching between
		// experiments. users still pinned to the previous experiment will see
		// failures because the old experiment name is unknown.
		if err != nil {
			log.Printf("repinned after error: %s", err.Error())
			selected := e.Select()
			return selected, makeTimestampedTag(selected, now), nil
		}

		return v, makeTimestampedTag(v, ts), err
	}

	selected := e.Select()
	return selected, makeTimestampedTag(selected, now), nil
}

// GetVariation selects the appropriate variation given it's 1 indexed ordinal
func (e *Experiment) GetVariation(ordinal int) (Variation, error) {
	if l := len(e.Variations); ordinal < 0 || ordinal > l {
		return Variation{}, fmt.Errorf("ordinal %d not in [1,%d]", ordinal, l)
	}

	return e.Variations[ordinal-1], nil
}

// GetTaggedVariation selects the appropriate variation given it's tag
func (e *Experiment) GetTaggedVariation(tag string) (Variation, error) {
	for _, variation := range e.Variations {
		if variation.Tag == tag {
			return variation, nil
		}
	}

	return Variation{}, fmt.Errorf("tag '%s' is not in experiment %s", tag, e.Name)
}

// makeTimestampedTag returns the variation tag as <tag>:<timestampNow>
func makeTimestampedTag(v Variation, now int64) string {
	return fmt.Sprintf("%s:%s", v.Tag, strconv.FormatInt(now, 10))
}

// Variation describes endpoints which are mapped onto strategy arms.
type Variation struct {
	Ordinal     int    // 1 indexed arm ordinal
	URL         string // the url associated with this variation, for out of band
	Tag         string // this tag is used throughout the lifecycle of the experiment
	Description string // freitext
}

// Variations is a set of variations sorted by ordinal.
type Variations []Variation

func (v Variations) Len() int           { return len(v) }
func (v Variations) Less(i, j int) bool { return v[i].Ordinal < v[j].Ordinal }
func (v Variations) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }

// NewExperiments reads in a json file and converts it to a map of experiments.
func NewExperiments(o Opener) (*Experiments, error) {
	file, err := o.Open()
	if err != nil {
		return &Experiments{}, fmt.Errorf("need a valid input file: %v", err)
	}

	defer file.Close()

	jsonString, err := ioutil.ReadAll(file)
	if err != nil {
		return &Experiments{}, fmt.Errorf("could not read jsony: %s", err.Error())
	}

	type variationConfig struct {
		URL         string `json:"url"`
		Description string `json:"description"`
		Ordinal     int    `json:"ordinal"`
	}

	type experimentsConfig struct {
		Name             string            `json:"experiment_name"`
		Strategy         string            `json:"strategy"`
		Snapshot         string            `json:"snapshot"`
		SnapshotPoll     int               `json:"snapshot-poll-seconds"`
		Parameters       []float64         `json:"parameters"`
		Variations       []variationConfig `json:"variations"`
		PreferredOrdinal int               `json:"preferred"`
	}

	var cfg []experimentsConfig
	if err := json.Unmarshal(jsonString, &cfg); err != nil {
		return &Experiments{}, fmt.Errorf("could not marshal json: %s ", err.Error())
	}

	// have to specify poll duration along with snapshot location
	for _, c := range cfg {
		if c.Snapshot != "" && c.SnapshotPoll == 0 {
			return &Experiments{}, fmt.Errorf("%s is missing snapshot-poll-seconds", c.Name)
		}
	}

	es := Experiments{}
	for _, e := range cfg {
		if e.PreferredOrdinal == 0 {
			return &Experiments{}, fmt.Errorf("could not make strategy: preferred variation missing")
		}

		strategy, err := New(len(e.Variations), e.Strategy, e.Parameters)
		if err != nil {
			return &Experiments{}, fmt.Errorf("could not make strategy: %s ", err.Error())
		}

		// this is a delayed strategy; gets it's internal state from a snapshot
		if e.Snapshot != "" {
			opener := NewOpener(e.Snapshot)
			duration := time.Duration(e.SnapshotPoll) * time.Second
			strategy, err = NewDelayed(strategy, opener, duration)
			if err != nil {
				return &Experiments{}, fmt.Errorf("could not delay strategy: %s ", err.Error())
			}
		}

		experiment := Experiment{
			Name:     e.Name,
			Strategy: strategy,
		}

		es[e.Name] = &experiment

		for _, v := range e.Variations {
			if v.Ordinal == e.PreferredOrdinal {
				experiment.PreferredOrdinal = v.Ordinal
			}

			experiment.Variations = append(experiment.Variations, Variation{
				Ordinal:     v.Ordinal,
				URL:         v.URL,
				Tag:         fmt.Sprintf("%s:%d", e.Name, v.Ordinal),
				Description: v.Description,
			})
		}

		if experiment.PreferredOrdinal == 0 {
			return &Experiments{}, fmt.Errorf("preferred variation ordinal %d not found in variations", e.PreferredOrdinal)
		}

		sort.Sort(experiment.Variations)
	}

	return &es, nil
}

// Experiments is an index of names to experiment
type Experiments map[string]*Experiment

// GetVariation returns the Experiment and variation pointed to by a string tag.
func (e *Experiments) GetVariation(tag string) (Experiment, Variation, error) {
	for _, experiment := range *e {
		for _, variation := range experiment.Variations {
			if variation.Tag == tag {
				return *experiment, variation, nil
			}
		}
	}

	return Experiment{}, Variation{}, fmt.Errorf("could not find variation '%s'", tag)
}

// TimestampedTagToTag docodes a timestamped tag in the form <tag>:<timestamp> into
// a (tag, ts)
func TimestampedTagToTag(timestampedTag string) (string, int64, error) {
	sep := strings.LastIndex(timestampedTag, ":")
	if sep == -1 {
		return "", 0, fmt.Errorf("invalid timestampedTag, does not end in :<timestamp>")
	}

	tag, at := timestampedTag[:sep], timestampedTag[sep+1:]
	ts, err := strconv.ParseInt(at, 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("invalid ttl: %s", err.Error())
	}

	return tag, ts, nil
}
