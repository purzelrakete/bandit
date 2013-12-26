package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// mapper returns a hadoop streaming mapper function. Emits (arm, reward)
// tuples onto the given writer, for the specified experiment only.
func mapper(s *Statistics, r io.Reader, w io.Writer) func() {
	return func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			for _, stat := range s.Stats {
				if key, value, ok := stat.mapLine(line); ok {
					fmt.Fprintf(w, "%s	%s\n", key, value)
				}
			}
		}
	}
}

// reducer returns a hadoop streaming reducer function. Emits one SnapshotLine
// for the specificed experiment.
func reducer(s *Statistics, r io.Reader, w io.Writer) func() {
	return func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			for _, stat := range s.Stats {
				stat.reduceLine(line)
			}
		}

		for _, stat := range s.Stats {
			if values, ok := stat.result(); ok {
				for key, value := range values {
					fmt.Fprintf(w, "%s	%d	%f\n", stat.getPrefix(), key+1, value)
				}
			}
		}
	}
}

// collector aggregates outputs of reducers
func collector(s *Statistics, r io.Reader, w io.Writer) func() {
	return func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			for _, stat := range s.Stats {
				stat.collect(line)
			}
		}

		fmt.Fprint(w, tsvSnapshot(s.rewards()), "\n")
	}
}

// tsvSnapshot is the tsv formatted snapshot file.
func tsvSnapshot(counts []int, rewards []float64) string {
	var values []string
	for _, reward := range rewards {
		values = append(values, fmt.Sprintf("%f", float64(reward)))
	}

	return strings.Join([]string{
		fmt.Sprintf("%d", len(rewards)),
		strings.Join(values, "\t"),
	}, "\t")
}
