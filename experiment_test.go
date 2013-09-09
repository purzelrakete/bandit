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
