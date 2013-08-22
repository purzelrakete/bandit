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
		counters := newCounters(arms)
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

			counters.Update(int(variant), reward)
		}

		fmt.Fprintln(w, SnapshotLine(counters, e))
	}
}

// SnapshotLine returns a snapshot log line
func SnapshotLine(c counters, e Experiment) string {
	var values []string
	for _, reward := range c.values {
		values = append(values, fmt.Sprintf("%f", float64(reward)))
	}

	return strings.Join([]string{
		e.Name,
		fmt.Sprintf("%d", len(e.Variants)),
		strings.Join(values, " "),
	}, " ")
}

// ParseSnapshot reads in a snapshot file. Snapshot files contain a single
// line per experiment, for example:
//
// shape-20130822 2 0.1 0.5
// shape-20130317 3 0.111 0.87 0.8901
//
// Tokens are separated by whitespace. The given example encodes an experiment
// with two variants. The experiment name is immediately followed by it's
// number of variants. This is followed by rewards (mean reward for each arm).
func ParseSnapshot(s io.Reader, campaign string) (counters, error) {
	var found bool
	var c counters
	scanner := bufio.NewScanner(s)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Index(line, campaign) == 0 {
			if found {
				return counters{}, fmt.Errorf("> 1 campaign snapshot: %s", campaign)
			}

			found = true
			fields := strings.Fields(line)
			arms, err := strconv.ParseInt(fields[1], 10, 16)
			if err != nil {
				return counters{}, fmt.Errorf("arms not an int: %s", err.Error())
			}

			if int(arms) != len(fields)-2 {
				return counters{}, fmt.Errorf("more fields than arms.")
			}

			var rewards []float64
			for _, str := range fields[2:] {
				reward, err := strconv.ParseFloat(str, 64)
				if err != nil {
					return counters{}, fmt.Errorf("rewards malformed: %s", err.Error())
				}

				rewards = append(rewards, reward)
			}

			c = newCounters(int(arms))
			c.values = rewards
		}
	}

	return c, nil
}
