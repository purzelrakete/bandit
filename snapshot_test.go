package bandit

import (
	"strings"
	"testing"
)

func TestParseSnapshot(t *testing.T) {
	input := strings.NewReader("2	0.120000	0.300000")

	s, err := ParseSnapshot(input)
	if err != nil {
		t.Fatalf("could not parse snapshot file: %s", err)
	}

	expectedArms := 2
	if got := s.arms; got != expectedArms {
		t.Fatalf("expected %d arms but got %d", expectedArms, got)
	}

	expectedReward := float64(0.12)
	if got := s.values[0]; got != expectedReward {
		t.Fatalf("expected arms to be %f but got %f", expectedReward, got)
	}

	expectedReward = float64(0.3)
	if got := s.values[1]; got != expectedReward {
		t.Fatalf("expected arms to be %f but got %f", expectedReward, got)
	}
}
