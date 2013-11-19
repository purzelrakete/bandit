package main

import (
	"bytes"
	"fmt"
	"github.com/purzelrakete/bandit"
	"log"
	"os"
	"strings"
	"time"
)

// Simple produces a snapshot every `poll` duration. FIXME: O(N) memory
func simple(s *bandit.Statistics, logFile string, poll time.Duration) error {
	snapshotFile := s.ExperimentName + ".tsv"
	opener := bandit.NewOpener(logFile)
	file, err := opener.Open()
	if err != nil {
		return fmt.Errorf("could not open logs: %s", err.Error())
	}

	defer file.Close()
	go func() {
		t := time.NewTicker(poll)
		for _ = range t.C {
			file, err := opener.Open()
			if err != nil {
				log.Printf("error opening log: %s", err.Error())
			}

			r, w := file, new(bytes.Buffer)
			mapper := bandit.SnapshotMapper(s, r, w)
			mapper()
			mapped := w.String()

			rS, w := strings.NewReader(mapped), new(bytes.Buffer)
			reducer := bandit.SnapshotReducer(s, rS, w)

			reducer()
			reduced := strings.TrimRight(w.String(), "\n ")

			snapshot, err := os.Create(snapshotFile)
			if err != nil {
				log.Printf("error creating snapshot file: %s", err.Error())
			}

			defer snapshot.Close()
			snapshot.Write([]byte(reduced))
		}
	}()

	select {}
}
