// Copyright 2013 SoundCloud, Rany Keddo. All rights reserved.  Use of this
// source code is governed by a license that can be found in the LICENSE file.

package bandit

import "testing"

func TestExperiment(t *testing.T) {
	es, err := NewExperiments("experiments.tsv")
	if err != nil {
		t.Fatalf("while reading experiment fixture: %s", err.Error())
	}

	expected := 2
	if got := len((*es)["shape-20130822"].Variants); got != expected {
		t.Fatalf("expected %d variants, got %d", expected, got)
	}
}

func TestPinToTag(t *testing.T) {
	tag, ts, err := PinToTag("shape-20130822:c8-circle:1378823906")
	if err != nil {
		t.Fatal("failed to parse pin: %s", err.Error())
	}

	if expected := "shape-20130822:c8-circle"; tag != expected {
		t.Fatalf("expected %s but got %s", expected, tag)
	}

	if expected := int64(1378823906); ts != expected {
		t.Fatalf("expected %d but got %d", expected, ts)
	}
}
