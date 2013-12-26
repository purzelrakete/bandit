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
func simple(s *Statistics, logFile string, poll time.Duration) error {
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

			// map
			rM, wM := file, new(bytes.Buffer)
			m := mapper(s, rM, wM)
			m()
			mapped := wM.String()

			// reduce
			rR, wR := strings.NewReader(mapped), new(bytes.Buffer)
			r := reducer(s, rR, wR)
			r()
			reduced := strings.TrimRight(wR.String(), "\n ")

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
