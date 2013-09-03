// Copyright 2013 SoundCloud, Rany Keddo. All rights reserved. Use of this
// source code is governed by a license that can be found in the LICENSE file.

package bandit

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
)

// SnapshotMapper returns a hadoop streaming mapper function. Emits (arm,
// reward) tuples onto the given writer, for the specified experiment only.
func SnapshotMapper(e Experiment, r io.Reader, w io.Writer) func() {
	reward := banditReward + " " + e.Name
	rewardLen := 7

	return func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Index(line, reward) >= 0 {
				fields := strings.Fields(line)
				if len(fields) != rewardLen {
					log.Fatalf("line does not have %d fields: '%s'", rewardLen, line)
				}

				variant, err := GetTaggedVariant(e, fields[5])
				if err != nil {
					log.Fatalf("invalid variant on line '%s': %s", line, err.Error())
				}

				reward, err := strconv.ParseFloat(fields[6], 32)
				if err != nil {
					log.Fatalf("non-float reward on line '%s': %s", line, err.Error())
				}

				fmt.Fprintf(w, "%d %f\n", variant.Ordinal, reward)
			}
		}
	}
}

// SnapshotReducer returns a hadoop streaming reducer function. Emits one
// SnapshotLine for the specificed experiment.
func SnapshotReducer(e Experiment, r io.Reader, w io.Writer) func() {
	return func() {
		arms := len(e.Variants)
		c := NewCounters(arms)
		scanner := bufio.NewScanner(r)

		for scanner.Scan() {
			line := scanner.Text()
			fields := strings.Fields(line)

			variant, err := strconv.ParseInt(fields[0], 10, 16)
			if err != nil {
				log.Fatalf("non-integral arm on line '%s': %s", line, err.Error())
			}

			reward, err := strconv.ParseFloat(fields[1], 32)
			if err != nil {
				log.Fatalf("non-float reward on line '%s': %s", line, err.Error())
			}

			c.Update(int(variant), reward)
		}

		fmt.Fprintln(w, SnapshotLine(c))
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
		return Counters{}, fmt.Errorf("more fields than arms.")
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
