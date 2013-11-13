// Copyright 2013 SoundCloud, Rany Keddo. All rights reserved.  Use of this
// source code is governed by a license that can be found in the LICENSE file.

package bandit

import "testing"

func TestExperiment(t *testing.T) {
	es, err := NewExperiments(NewFileOpener("experiments.tsv"))
	if err != nil {
		t.Fatalf("while reading experiment fixture: %s", err.Error())
	}

	e, ok := (*es)["shape-20130822"]
	if !ok {
		t.Fatalf("could not find test campaign")
	}

	expected := 2
	if got := len(e.Variations); got != expected {
		t.Fatalf("expected %d variations, got %d", expected, got)
	}

	expectedTag := "shape-20130822:1"
	if got := e.Variations[0].Tag; got != expectedTag {
		t.Fatalf("expected variation tag %s, got %s", expectedTag, got)
	}
}

func TestTimestampedTagToTag(t *testing.T) {
	tag, ts, err := TimestampedTagToTag("shape-20130822:c8-circle:1378823906")
	if err != nil {
		t.Fatalf("failed to parse timstamped tag: %s", err.Error())
	}

	if expected := "shape-20130822:c8-circle"; tag != expected {
		t.Fatalf("expected %s but got %s", expected, tag)
	}

	if expected := int64(1378823906); ts != expected {
		t.Fatalf("expected %d but got %d", expected, ts)
	}
}
