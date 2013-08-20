// Copyright 2013 SoundCloud, Rany Keddo. All rights reserved.  Use of this
// source code is governed by a license that can be found in the LICENSE file.

package bandit

import "testing"

func TestCampaign(t *testing.T) {
	campaigns, err := ParseCampaigns("campaigns.tsv")
	if err != nil {
		t.Fatalf("while reading campaign fixture: %s", err.Error())
	}

	expected := 3
	if got := len(campaigns["widgets"].Variants); got != expected {
		t.Fatalf("expected %d variants, got %d", expected, got)
	}
}
