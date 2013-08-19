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
