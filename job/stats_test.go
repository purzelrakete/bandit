package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestMapper(t *testing.T) {
	log := []string{
		"1379069548	BanditSelection	shape-20130822:2:1",
		"1379069749	BanditSelection	shape-20130822:2:1",
		"1379069948	BanditSelection	plants-20121111:1:2",
		"1379069648	BanditReward	shape-20130822:2:1 1.0",
		"1379069848	BanditReward	shape-20130822:2:1 0.0",
		"1379069158	BanditReward	plants-20121111:1:2 1.0",
		"1379069258	BanditReward	plants-20121111:1:2 1.0",
	}

	stats := newStatistics("shape-20130822")

	r, w := strings.NewReader(strings.Join(log, "\n")), new(bytes.Buffer)
	mapper := mapper(stats, r, w)

	mapper()
	mapped := strings.TrimRight(w.String(), "\n ")

	expected := strings.Join([]string{
		"BanditSelection_2	1",
		"BanditSelection_2	1",
		"BanditReward_2	1.0",
		"BanditReward_2	0.0",
	}, "\n")

	if got := mapped; got != expected {
		t.Fatalf("expected '%s' but got '%s'", expected, got)
	}
}

func TestReducer(t *testing.T) {
	log := []string{
		"BanditSelection_1	1",
		"BanditSelection_1	1",
		"BanditSelection_2	1",
		"BanditSelection_2	1",
		"BanditReward_1	1.0",
		"BanditReward_1	0.0",
	}

	stats := newStatistics("shape-20130822")

	r, w := strings.NewReader(strings.Join(log, "\n")), new(bytes.Buffer)
	reducer := reducer(stats, r, w)

	reducer()
	reduced := strings.TrimRight(w.String(), "\n ")

	expected := strings.Join([]string{
		"BanditReward	1	1.000000",
		"BanditSelection	1	2.000000",
		"BanditSelection	2	2.000000",
	}, "\n")

	if got := reduced; got != expected {
		t.Fatalf("expected '%s' but got '%s'", expected, got)
	}
}

func TestMapperReducer(t *testing.T) {
	log := []string{
		"1379069548	BanditSelection	shape-20130822:2:1",
		"1379069749	BanditSelection	shape-20130822:2:2",
		"1379069948	BanditSelection	plants-20121111:1:3",
		"1379069648	BanditReward	shape-20130822:2:1	1.0",
		"1379069848	BanditReward	shape-20130822:2:2	0.0",
		"1379069158	BanditReward	plants-20121111:1:3	1.0",
		"1379069258	BanditReward	plants-20121111:1:3	1.0",
	}

	stats := newStatistics("shape-20130822")

	r, w := strings.NewReader(strings.Join(log, "\n")), new(bytes.Buffer)
	mapper := mapper(stats, r, w)

	mapper()
	mapped := w.String()

	r, w = strings.NewReader(mapped), new(bytes.Buffer)

	reducer := reducer(stats, r, w)

	reducer()
	reduced := strings.TrimRight(w.String(), "\n ")

	expected := strings.Join([]string{
		"BanditReward	2	1.000000",
		"BanditSelection	2	2.000000",
	}, "\n")

	if got := reduced; got != expected {
		t.Fatalf("expected '%s' but got '%s'", expected, got)
	}
}

func TestCollect(t *testing.T) {
	log := []string{
		"BanditReward	2	1.000000",
		"BanditSelection	2	2.000000",
		"BanditReward	1	2.000000",
		"BanditSelection	1	4.000000",
	}

	stats := newStatistics("shape-20130822")

	r, w := strings.NewReader(strings.Join(log, "\n")), new(bytes.Buffer)
	collect := collector(stats, r, w)
	collect()
	collected := strings.TrimRight(w.String(), "\n ")

	expected := "2	0.500000	0.500000"

	if got := collected; got != expected {
		t.Fatalf("expected '%s' but got '%s'", expected, got)
	}
}

func TestSnapshotCounter(t *testing.T) {
	log := []string{
		"BanditReward	2	1.000000",
		"BanditSelection	2	4.000000",
		"BanditReward	1	2.000000",
		"BanditSelection	1	4.000000",
	}

	stats := newStatistics("shape-20130822")
	r, w := strings.NewReader(strings.Join(log, "\n")), new(bytes.Buffer)

	collect := collector(stats, r, w)
	collect()
	counts, rewards := stats.rewards()

	expected := "2	0.500000	0.250000"
	snapshot := tsvSnapshot(counts, rewards)

	if got := snapshot; got != expected {
		t.Fatalf("expected '%s' but got '%s'", expected, got)
	}
}
