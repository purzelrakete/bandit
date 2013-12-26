package main

// FIXME: try to remove almost all of this code. No need for interfaces and
// general code just to produce mean rewards.

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

const (
	banditSelection = "BanditSelection"
	banditReward    = "BanditReward"
)

// statistics contains all stats which should be computed
type statistics struct {
	experimentName string
	stats          []stats
}

// newStatistics creates a new object with default statistics
func newStatistics(experimentName string) *statistics {
	return &statistics{
		experimentName: experimentName,
		stats: []stats{
			newSumRewards(experimentName),
			newCountSelects(experimentName),
		},
	}
}

func (s *statistics) rewards() ([]int, []float64) {
	rewards, ok := s.stats[0].result()
	if !ok {
		panic("no rewards")
	}

	selects, ok := s.stats[1].result()
	if !ok {
		panic("no selects")
	}

	if len(selects) != len(rewards) {
		panic("rewards, selects have different number of arms")
	}

	rCounts := make([]int, len(rewards))
	rRewards := make([]float64, len(rewards))
	for key := range rewards {
		index := key - 1
		rCounts[index] = int(selects[key])
		rRewards[index] = rewards[key] / selects[key]
	}

	return rCounts, rRewards
}

// stats aggregates statistics from line based input
type stats interface {
	mapLine(string) (string, string, bool) // line -> (key, value, matches)
	reduceLine(string)
	result() (map[int]float64, bool)
	collect(string)
	getPrefix() string
}

type countSelects struct {
	selects        map[int]float64
	prefix         string
	experimentName string
}

func newCountSelects(name string) stats {
	return &countSelects{
		prefix:         "BanditSelection",
		experimentName: name,
		selects:        make(map[int]float64),
	}
}

func (c *countSelects) getPrefix() string {
	return c.prefix
}

// mapLine to count selects from a log file
func (c *countSelects) mapLine(line string) (string, string, bool) {
	selection := banditSelection + "\t" + c.experimentName
	selectionLen := 3
	if strings.Index(line, selection) >= 0 {
		fields := strings.Fields(line)
		if len(fields) != selectionLen {
			log.Fatalf("line does not have %d fields: '%s'", selectionLen, line)
		}

		splittedString := strings.Split(fields[2], ":")
		variation, err := strconv.ParseInt(splittedString[1], 10, 0)
		if err != nil {
			log.Fatalf("invalid variation in line '%s': %s", line, err.Error())
		}

		return fmt.Sprintf("%s_%d", c.prefix, variation), "1", true
	}

	return "", "", false
}

// reduceLine to sum the selects per arm
func (c *countSelects) reduceLine(line string) {
	if strings.Index(line, c.prefix) >= 0 {
		preparedString := strings.Replace(line, "_", "\t", 1)
		fields := strings.Fields(preparedString)
		variation, err := strconv.Atoi(fields[1])
		if err != nil {
			log.Fatalf("non-integral arm on line '%s': %s", line, err.Error())
		}
		c.selects[variation-1]++
	}
}

func (c *countSelects) collect(line string) {
	if strings.Index(line, c.prefix) >= 0 {
		fields := strings.Fields(line)
		variation, err := strconv.Atoi(fields[1])
		if err != nil {
			log.Fatalf("non-integral arm on line '%s': %s", line, err.Error())
		}
		selects, err := strconv.ParseFloat(fields[2], 32)
		if err != nil {
			log.Fatalf("non-float selects on line '%s': %s", line, err.Error())
		}
		c.selects[variation] = selects
	}
}

func (c *countSelects) result() (map[int]float64, bool) {
	if len(c.selects) > 0 {
		return c.selects, true
	}
	return map[int]float64{}, false
}

// sumRewards to sum rewards from log lines
type sumRewards struct {
	prefix         string
	experimentName string
	rewards        map[int]float64
}

func newSumRewards(name string) stats {
	return &sumRewards{
		prefix:         "BanditReward",
		experimentName: name,
		rewards:        make(map[int]float64),
	}
}

func (s *sumRewards) getPrefix() string {
	return s.prefix
}

// mapLine mapper emmits a key, value for each Reward line in log file
func (s *sumRewards) mapLine(line string) (string, string, bool) {
	reward := banditReward + "\t" + s.experimentName
	rewardLen := 4
	if strings.Index(line, reward) >= 0 {
		fields := strings.Fields(line)
		if len(fields) != rewardLen {
			log.Fatalf("line does not have %d fields: '%s'", rewardLen, line)
		}

		splittedString := strings.Split(fields[2], ":")
		variation, err := strconv.ParseInt(splittedString[1], 10, 0)
		if err != nil {
			log.Fatalf("invalid variation on line '%s': %s", line, err.Error())
		}

		return fmt.Sprintf("%s_%d", s.prefix, variation), fields[3], true
	}
	return "", "", false
}

// reduceLine reducer sums up the incomming rewards
func (s *sumRewards) reduceLine(line string) {
	if strings.Index(line, s.prefix) >= 0 {
		preparedString := strings.Replace(line, "_", "\t", 1)
		fields := strings.Fields(preparedString)
		variation, err := strconv.Atoi(fields[1])
		if err != nil {
			log.Fatalf("non-integral arm on line '%s': %s", line, err.Error())
		}

		reward, err := strconv.ParseFloat(fields[2], 32)
		if err != nil {
			log.Fatalf("non-float reward on line '%s': %s", line, err.Error())
		}

		s.rewards[variation-1] += reward
	}
}

func (s *sumRewards) result() (map[int]float64, bool) {
	if len(s.rewards) > 0 {
		return s.rewards, true
	}
	return map[int]float64{}, false
}

func (s *sumRewards) collect(line string) {
	if strings.Index(line, s.prefix) >= 0 {
		fields := strings.Fields(line)
		variation, err := strconv.Atoi(fields[1])
		if err != nil {
			log.Fatalf("non-integral arm on line '%s': %s", line, err.Error())
		}
		reward, err := strconv.ParseFloat(fields[2], 32)
		if err != nil {
			log.Fatalf("non-float reward on line '%s': %s", line, err.Error())
		}
		s.rewards[variation] = reward
	}
}
