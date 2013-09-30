// Copyright 2013 SoundCloud, Rany Keddo. All rights reserved. Use of this
// source code is governed by a license that can be found in the LICENSE file.

package bandit

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// SnapshotMapper returns a hadoop streaming mapper function. Emits (arm,
// reward) tuples onto the given writer, for the specified experiment only.
func SnapshotMapper(experimentName string, s []Stats, r io.Reader, w io.Writer) func() {
	return func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			for _, stat := range s {
				if key, value, ok := stat.mapLine(line); ok {
					fmt.Fprintf(w, "%s %s\n", key, value)
				}
			}
		}
	}
}

// SnapshotReducer returns a hadoop streaming reducer function. Emits one
// SnapshotLine for the specificed experiment.
func SnapshotReducer(experimentName string, s []Stats, r io.Reader, w io.Writer) func() {
	return func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			for _, stat := range s {
				stat.reduceLine(line)
			}
		}

		for _, stat := range s {
			if values, ok := stat.result(); ok {
				for key, value := range values {
					fmt.Fprintf(w, "%s %d %f\n", stat.getPrefix(), key+1, value)
				}
			}
		}
	}
}

// SnapshotCollect
func SnapshotCollect(e *Experiment, s []Stats, r io.Reader, w io.Writer) func() {
	return func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			for _, stat := range s {
				stat.collect(line)
			}
		}

		for _, stat := range s {
			if values, ok := stat.result(); ok {
				fmt.Fprintf(w, "%s %d", stat.getPrefix(), len(values))
				for key, value := range values {
					fmt.Fprintf(w, " %d:%f", key, value)
				}
				fmt.Fprintf(w, "\n")
			}
		}
	}
}

// SnapshotLine returns a snapshot log line
func SnapshotLine(c Counters) string {
	var values []string
	for _, reward := range c.values {
		values = append(values, fmt.Sprintf("%f", float64(reward)))
	}

	return strings.Join([]string{
		fmt.Sprintf("%d", c.arms),
		strings.Join(values, " "),
	}, " ")
}

// ParseSnapshot reads in a snapshot file. Snapshot files contain a single
// line experiment snapshot, for example:
//
// 2 0.1 0.5
//
// Tokens are separated by whitespace. The given example encodes an experiment
// with two variants. First is the number of variants. This is followed by
// rewards (mean reward for each arm).
func ParseSnapshot(s io.Reader) (Counters, error) {
	lines := 0
	var line string
	for scanner := bufio.NewScanner(s); scanner.Scan(); lines++ {
		if lines > 1 {
			return Counters{}, fmt.Errorf("> 1 line in snapshot")
		}

		line = scanner.Text()
	}

	fields := strings.Fields(line)
	arms, err := strconv.ParseInt(fields[0], 10, 16)
	if err != nil {
		return Counters{}, fmt.Errorf("arms not an int: %s", err.Error())
	}

	if int(arms) != len(fields)-1 {
		return Counters{}, fmt.Errorf("more fields than arms")
	}

	var rewards []float64
	for _, str := range fields[1:] {
		reward, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return Counters{}, fmt.Errorf("rewards malformed: %s", err.Error())
		}

		rewards = append(rewards, reward)
	}

	c := NewCounters(int(arms))
	c.values = rewards

	return c, nil
}

// GetSnapshot returns Counters given a snapshot filename.
func GetSnapshot(o Opener) (Counters, error) {
	reader, err := o.Open()
	if err != nil {
		return Counters{}, fmt.Errorf("could not open: %s", err.Error())
	}

	counters, err := ParseSnapshot(reader)
	if err != nil {
		return Counters{}, fmt.Errorf("could not parse snapshot: %s", err.Error())
	}

	return counters, nil
}
