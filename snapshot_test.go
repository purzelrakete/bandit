package bandit

import (
	"bytes"
	"strings"
	"testing"
)

func TestSnapshotMapper(t *testing.T) {
	log := []string{
		"1379069548 BanditSelection shape-20130822:2",
		"1379069749 BanditSelection shape-20130822:2",
		"1379069948 BanditSelection plants-20121111:1",
		"1379069648 BanditReward shape-20130822:2 1.0",
		"1379069848 BanditReward shape-20130822:2 0.0",
		"1379069158 BanditReward plants-20121111:1 1.0",
		"1379069258 BanditReward plants-20121111:1 1.0",
	}

	stats := NewStatistics("shape-20130822")

	r, w := strings.NewReader(strings.Join(log, "\n")), new(bytes.Buffer)
	mapper := SnapshotMapper("shape-20130822", stats, r, w)

	mapper()
	mapped := w.String()

	expected := strings.Join([]string{
		"BanditSelection_2 1",
		"BanditSelection_2 1",
		"BanditReward_2 1.0",
		"BanditReward_2 0.0",
		"",
	}, "\n")

	if got := mapped; got != expected {
		t.Fatalf("expected '%s' but got '%s'", expected, got)
	}
}

func TestSnapshotReducer(t *testing.T) {
	log := []string{
		"BanditSelection_1 1",
		"BanditSelection_1 1",
		"BanditSelection_2 1",
		"BanditSelection_2 1",
		"BanditReward_1 1.0",
		"BanditReward_1 0.0",
		"",
	}

	stats := NewStatistics("shape-20130822")

	r, w := strings.NewReader(strings.Join(log, "\n")), new(bytes.Buffer)
	reducer := SnapshotReducer("shape-20130822", stats, r, w)

	reducer()
	reduced := strings.TrimRight(w.String(), "\n ")

	expected := strings.Join([]string{
		"BanditReward 1 1.000000",
		"BanditSelection 1 2.000000",
		"BanditSelection 2 2.000000",
	}, "\n")

	if got := reduced; got != expected {
		t.Fatalf("expected '%s' but got '%s'", expected, got)
	}
}

func TestSnapshotMapperReducer(t *testing.T) {
	log := []string{
		"1379069548 BanditSelection shape-20130822:2",
		"1379069749 BanditSelection shape-20130822:2",
		"1379069948 BanditSelection plants-20121111:1",
		"1379069648 BanditReward shape-20130822:2 1.0",
		"1379069848 BanditReward shape-20130822:2 0.0",
		"1379069158 BanditReward plants-20121111:1 1.0",
		"1379069258 BanditReward plants-20121111:1 1.0",
	}

	stats := NewStatistics("shape-20130822")

	r, w := strings.NewReader(strings.Join(log, "\n")), new(bytes.Buffer)
	mapper := SnapshotMapper("shape-20130822", stats, r, w)

	mapper()
	mapped := w.String()

	r, w = strings.NewReader(mapped), new(bytes.Buffer)

	reducer := SnapshotReducer("shape-20130822", stats, r, w)

	reducer()
	reduced := strings.TrimRight(w.String(), "\n ")

	expected := strings.Join([]string{
		"BanditReward 2 1.000000",
		"BanditSelection 2 2.000000",
	}, "\n")

	if got := reduced; got != expected {
		t.Fatalf("expected '%s' but got '%s'", expected, got)
	}
}

func TestSnapshotCollect(t *testing.T) {
	log := []string{
		"BanditReward 2 1.000000",
		"BanditSelection 2 2.000000",
		"BanditReward 1 2.000000",
		"BanditSelection 1 4.000000",
	}

	stats := NewStatistics("shape-20130822")

	r, w := strings.NewReader(strings.Join(log, "\n")), new(bytes.Buffer)
	collect := SnapshotCollect(stats, r, w)
	collect()
	collected := strings.TrimRight(w.String(), "\n ")

	expected := "2 0.500000 0.500000"

	if got := collected; got != expected {
		t.Fatalf("expected '%s' but got '%s'", expected, got)
	}

}

func TestSnapshotCounter(t *testing.T) {
	log := []string{
		"BanditReward 2 1.000000",
		"BanditSelection 2 4.000000",
		"BanditReward 1 2.000000",
		"BanditSelection 1 4.000000",
	}

	stats := NewStatistics("shape-20130822")

	r, w := strings.NewReader(strings.Join(log, "\n")), new(bytes.Buffer)
	collect := SnapshotCollect(stats, r, w)
	collect()
	counters := stats.getCounters()

	expected := "2 0.500000 0.250000"

	snapshot := SnapshotLine(counters)

	if got := snapshot; got != expected {
		t.Fatalf("expected '%s' but got '%s'", expected, got)
	}

}

func TestParseSnapshot(t *testing.T) {
	input := strings.NewReader("2 0.120000 0.300000")

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
