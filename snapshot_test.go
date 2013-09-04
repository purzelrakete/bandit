package bandit

import (
	"bytes"
	"strings"
	"testing"
)

func TestSnapshot(t *testing.T) {
	log := []string{
		"2013/08/22 14:20:05 BanditSelection shape-20130822 0 shape-20130822:c8-circle",
		"2013/08/22 14:20:06 BanditReward shape-20130822 0 shape-20130822:c8-circle 1.0",
		"2013/08/22 14:20:07 BanditSelection shape-20130822 0 shape-20130822:c8-circle",
		"2013/08/22 14:20:08 BanditReward shape-20130822 0 shape-20130822:c8-circle 0.0",
		"2013/08/22 14:20:09 BanditSelection plants-20121111 0 plants-20121111:f1-camelia",
		"2013/08/22 14:20:10 BanditReward plant-20121111 0 plants-20121111:f1-camelia 1.0",
	}

	campaigns, err := ParseExperiments("experiments.tsv")
	if err != nil {
		t.Fatalf("while reading campaign fixture: %s", err.Error())
	}

	c, ok := campaigns["shape-20130822"]
	if !ok {
		t.Fatalf("could not find shapes campaign.")
	}

	r, w := strings.NewReader(strings.Join(log, "\n")), new(bytes.Buffer)
	mapper := SnapshotMapper(c, r, w)
	mapper()
	mapped := w.String()

	r, w = strings.NewReader(mapped), new(bytes.Buffer)
	reducer := SnapshotReducer(c, r, w)
	reducer()
	reduced := strings.TrimRight(w.String(), "\n ")

	expected := "2 0.000000 0.500000"
	if got := reduced; got != expected {
		t.Fatalf("expected '%s' but got '%s'", expected, got)
	}
}

func TestParseSnapshot(t *testing.T) {
	input := strings.NewReader("2 0.120000 0.300000")

	s, err := ParseSnapshot(input)
	if err != nil {
		t.Fatalf("could not parse snapshot file: ", err.Error())
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
