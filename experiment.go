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
	Name       string
	Bandit     Bandit
	Variations Variations
}

// Select calls SelectArm on the bandit and returns the associated variation
func (e *Experiment) Select() Variation {
	selected := e.Bandit.SelectArm()
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

// InitDelayedBandit adds a delayed bandit to this experiment.
func (e *Experiment) InitDelayedBandit(o Opener, poll time.Duration) error {
	if _, err := GetSnapshot(o); err != nil { // try once
		fmt.Errorf("could not get snapshot: %s", err.Error())
	}

	c := make(chan Counters)
	go func() {
		t := time.NewTicker(poll)
		for _ = range t.C {
			counters, err := GetSnapshot(o)
			if err != nil {
				log.Printf("BanditError: could not get snapshot: %s", err.Error())
			}

			c <- counters
		}
	}()

	b, _ := NewSoftmax(len(e.Variations), 0.1) // 0.1 cannot return an error
	d, err := NewDelayedBandit(b, c)
	if err != nil {
		return err
	}

	e.Bandit = d
	return nil
}

// Variation describes endpoints which are mapped onto bandit arms.
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
		Name       string            `json:"experiment_name"`
		Bandit     string            `json:"bandit"`
		Parameters []float64         `json:"parameters"`
		Variations []variationConfig `json:"variations"`
	}

	var cfg []experimentsConfig
	if err := json.Unmarshal(jsonString, &cfg); err != nil {
		return &Experiments{}, fmt.Errorf("could not marshal json: %s ", err)
	}

	es := Experiments{}
	for _, e := range cfg {
		bandit, _ := NewSoftmax(len(e.Variations), float64(0.1))
		experiment := Experiment{
			Name:   e.Name,
			Bandit: bandit,
		}

		es[e.Name] = &experiment

		for _, v := range e.Variations {
			experiment.Variations = append(experiment.Variations, Variation{
				Ordinal:     v.Ordinal,
				URL:         v.URL,
				Tag:         fmt.Sprintf("%s:%d", e.Name, v.Ordinal),
				Description: v.Description,
			})
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

// InitDelayedBandit initializes all bandits with delayed Softmax(0.1).
func (e *Experiments) InitDelayedBandit(o Opener, poll time.Duration) error {
	for _, e := range *e {
		if err := e.InitDelayedBandit(o, poll); err != nil {
			return fmt.Errorf("delayed bandit setup failed: %s", err.Error())
		}
	}

	return nil
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
