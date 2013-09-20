// Copyright 2013 SoundCloud, Rany Keddo. All rights reserved.  Use of this
// source code is governed by a license that can be found in the LICENSE file.

package bandit

import (
	"encoding/csv"
	"fmt"
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

// Experiment is a single experiment. Variants are in ascending ordinal
// sorting, where ordinals are contiguous and start at 1.
type Experiment struct {
	Name     string
	Bandit   Bandit
	Variants Variants
}

// Select calls SelectArm on the bandit and returns the associated variant
func (e *Experiment) Select() Variant {
	selected := e.Bandit.SelectArm()
	if selected > len(e.Variants) {
		panic("selected impossible arm")
	}

	v, _ := e.GetVariant(selected)
	return v
}

// SelectTimestamped selects the appropriate variant given it's
// timestampedTag. A timestamped tag is a string in the form
// <tag>:<timestamp>. If the duration between <timestamp> and the current time
// is smaller than `d`, the given tagged is used to return a variant. If it is
// larger, Select() is called instead.  If the `timestampedTag` argument is
// the blank string, Select() is called instead.
func (e *Experiment) SelectTimestamped(
	timestampedTag string,
	ttl time.Duration) (Variant, string, error) {
	now := time.Now().Unix()

	if timestampedTag == "" {
		selected := e.Select()
		return selected, makeTimestampedTag(selected, now), nil
	}

	tag, ts, err := TimestampedTagToTag(timestampedTag)
	if err != nil {
		return Variant{}, "", fmt.Errorf("bad timestamped tag: %s", err.Error())
	}

	// return the given timestamped tag
	if ttl > time.Since(time.Unix(ts, 0)) {
		v, err := e.GetTaggedVariant(tag)
		return v, makeTimestampedTag(v, ts), err
	}

	selected := e.Select()
	return selected, makeTimestampedTag(selected, now), nil
}

// GetVariant selects the appropriate variant given it's 1 indexed ordinal
func (e *Experiment) GetVariant(ordinal int) (Variant, error) {
	if l := len(e.Variants); ordinal < 0 || ordinal > l {
		return Variant{}, fmt.Errorf("ordinal %d not in [1,%d]", ordinal, l)
	}

	return e.Variants[ordinal-1], nil
}

// GetTaggedVariant selects the appropriate variant given it's tag
func (e *Experiment) GetTaggedVariant(tag string) (Variant, error) {
	for _, variant := range e.Variants {
		if variant.Tag == tag {
			return variant, nil
		}
	}

	return Variant{}, fmt.Errorf("tag '%s' is not in experiment %s", tag, e.Name)
}

// makeTimestampedTag returns the variant tag as <tag>:<timestampNow>
func makeTimestampedTag(v Variant, now int64) string {
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

	b, _ := NewSoftmax(len(e.Variants), 0.1) // 0.1 cannot return an error
	d, err := NewDelayedBandit(b, c)
	if err != nil {
		return err
	}

	e.Bandit = d
	return nil
}

// Variant describes endpoints which are mapped onto bandit arms.
type Variant struct {
	Ordinal     int    // 1 indexed arm ordinal
	URL         string // the url associated with this variant, for out of band
	Tag         string // this tag is used throughout the lifecycle of the experiment
	Description string // freitext
}

// Variants is a set of variants sorted by ordinal.
type Variants []Variant

func (v Variants) Len() int           { return len(v) }
func (v Variants) Less(i, j int) bool { return v[i].Ordinal < v[j].Ordinal }
func (v Variants) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }

// NewExperiments reads in a tsv file and converts it to a map of experiments.
func NewExperiments(o Opener) (*Experiments, error) {
	file, err := o.Open()
	if err != nil {
		return &Experiments{}, fmt.Errorf("need a valid input file: %v", err)
	}

	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = '\t'
	records, err := reader.ReadAll()
	if err != nil {
		return &Experiments{}, fmt.Errorf("could not read tsv: %s ", err)
	}

	// intermediary data structure groups variants
	type experimentVariants map[string]Variants

	variants := make(experimentVariants)
	for i, record := range records {
		if l := len(record); l != 4 {
			return &Experiments{}, fmt.Errorf("record is not %v long: %v", l, record)
		}

		ordinal, err := strconv.Atoi(record[1])
		if err != nil {
			return &Experiments{}, fmt.Errorf("invalid ordinal on line %d: %s", i, err)
		}

		name := record[0]
		if words := strings.Fields(name); len(words) != 1 {
			return &Experiments{}, fmt.Errorf("experiment has whitespace: %s", name)
		}

		description := record[3]
		variants[name] = append(variants[name], Variant{
			Ordinal:     ordinal,
			URL:         record[2],
			Tag:         fmt.Sprintf("%s:%s", name, record[1]),
			Description: description,
		})
	}

	// sorted experiment variants
	experiments := make(Experiments)
	for name, variants := range variants {
		sort.Sort(variants)
		b, _ := NewSoftmax(len(variants), 0.1) // default to softmax.
		experiments[name] = &Experiment{
			Bandit:   b,
			Name:     name,
			Variants: variants,
		}
	}

	// fail if ordinals are non-contiguous or do not start with 1
	for name, variants := range variants {
		for i := 0; i < len(variants); i++ {
			if ord := variants[i].Ordinal; ord != i+1 {
				return &Experiments{}, fmt.Errorf("%s: variant %d noncontiguous", name, ord)
			}
		}
	}

	return &experiments, nil
}

// Experiments is an index of names to experiment
type Experiments map[string]*Experiment

// GetVariant returns the Experiment and variant pointed to by a string tag.
func (e *Experiments) GetVariant(tag string) (Experiment, Variant, error) {
	for _, experiment := range *e {
		for _, variant := range experiment.Variants {
			if variant.Tag == tag {
				return *experiment, variant, nil
			}
		}
	}

	return Experiment{}, Variant{}, fmt.Errorf("could not find variant '%s'", tag)
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
