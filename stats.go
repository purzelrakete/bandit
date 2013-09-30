package bandit

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

// Stats aggregates statistics from line based input
type Stats interface {
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

func NewCountSelects(name string) Stats {
	return &countSelects{
		prefix:         "BanditSelection",
		experimentName: name,
		selects:        make(map[int]float64),
	}
}

func (c *countSelects) getPrefix() string {
	return c.prefix
}

// Mapper to count selects from a log file
func (c *countSelects) mapLine(line string) (string, string, bool) {
	selection := banditSelection + " " + c.experimentName
	selectionLen := 3
	if strings.Index(line, selection) >= 0 {
		fields := strings.Fields(line)
		if len(fields) != selectionLen {
			log.Fatalf("line does not have %d fields: '%s'", selectionLen, line)
		}

		splittedString := strings.Split(fields[2], ":")
		variant, err := strconv.ParseInt(splittedString[1], 10, 0)
		if err != nil {
			log.Fatalf("invalid variant in line '%s': %s", line, err.Error())
		}

		return fmt.Sprintf("%s_%d", c.prefix, variant), "1", true
	}
	return "", "", false
}

// Reducer to sum the selects per arm
func (c *countSelects) reduceLine(line string) {
	if strings.Index(line, c.prefix) >= 0 {
		preparedString := strings.Replace(line, "_", " ", 1)
		fields := strings.Fields(preparedString)
		variant, err := strconv.Atoi(fields[1])
		if err != nil {
			log.Fatalf("non-integral arm on line '%s': %s", line, err.Error())
		}
		c.selects[variant-1]++
	}
}

func (c *countSelects) collect(line string) {
	if strings.Index(line, c.prefix) >= 0 {
		fields := strings.Fields(line)
		variant, err := strconv.Atoi(fields[1])
		if err != nil {
			log.Fatalf("non-integral arm on line '%s': %s", line, err.Error())
		}
		selects, err := strconv.ParseFloat(fields[2], 32)
		if err != nil {
			log.Fatalf("non-float selects on line '%s': %s", line, err.Error())
		}
		c.selects[variant] = selects
	}
}

func (c *countSelects) result() (map[int]float64, bool) {
	if len(c.selects) > 0 {
		return c.selects, true
	}
	return map[int]float64{}, false
}

// structure with functions to sum rewards from log lines
type sumRewards struct {
	prefix         string
	experimentName string
	rewards        map[int]float64
}

func NewSumRewards(name string) Stats {
	return &sumRewards{
		prefix:         "BanditReward",
		experimentName: name,
		rewards:        make(map[int]float64),
	}
}

func (s *sumRewards) getPrefix() string {
	return s.prefix
}

// sumRewards mapper emmits a key, value for each Reward line in log file
func (s *sumRewards) mapLine(line string) (string, string, bool) {
	reward := banditReward + " " + s.experimentName
	rewardLen := 4
	if strings.Index(line, reward) >= 0 {
		fields := strings.Fields(line)
		if len(fields) != rewardLen {
			log.Fatalf("line does not have %d fields: '%s'", rewardLen, line)
		}

		splittedString := strings.Split(fields[2], ":")
		variant, err := strconv.ParseInt(splittedString[1], 10, 0)
		if err != nil {
			log.Fatalf("invalid variant on line '%s': %s", line, err.Error())
		}

		return fmt.Sprintf("%s_%d", s.prefix, variant), fields[3], true
	}
	return "", "", false
}

// sumRewards reducer sums up the incomming rewards
func (s *sumRewards) reduceLine(line string) {
	if strings.Index(line, s.prefix) >= 0 {
		preparedString := strings.Replace(line, "_", " ", 1)
		fields := strings.Fields(preparedString)
		variant, err := strconv.Atoi(fields[1])
		if err != nil {
			log.Fatalf("non-integral arm on line '%s': %s", line, err.Error())
		}

		reward, err := strconv.ParseFloat(fields[2], 32)
		if err != nil {
			log.Fatalf("non-float reward on line '%s': %s", line, err.Error())
		}

		s.rewards[variant-1] += reward
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
		variant, err := strconv.Atoi(fields[1])
		if err != nil {
			log.Fatalf("non-integral arm on line '%s': %s", line, err.Error())
		}
		reward, err := strconv.ParseFloat(fields[2], 32)
		if err != nil {
			log.Fatalf("non-float reward on line '%s': %s", line, err.Error())
		}
		s.rewards[variant] = reward
	}
}
