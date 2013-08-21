// Copyright 2013 SoundCloud, Rany Keddo. All rights reserved.  Use of this
// source code is governed by a license that can be found in the LICENSE file.

// Package bandit implements a multiarmed bandit. Runs A/B tests while
// controlling the tradeoff between exploring new arms and exploiting the
// currently best arm.
//
// The project is based on John Myles White's 'Bandit Algorithms for Website
// Optimization'.
package bandit

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// Bandit can select arm or update information
type Bandit interface {
	SelectArm() int
	Update(arm int, reward float64)
	Version() string
	Reset()
}

// NewEpsilonGreedy constructs an epsilon greedy bandit.
func NewEpsilonGreedy(arms int, epsilon float64) (Bandit, error) {
	if !(epsilon >= 0 && epsilon <= 1) {
		return &epsilonGreedy{}, fmt.Errorf("epsilon not in [0, 1]")
	}

	return &epsilonGreedy{
		counts:  make([]int, arms),
		values:  make([]float64, arms),
		rand:    rand.New(rand.NewSource(time.Now().UnixNano())),
		arms:    arms,
		epsilon: epsilon,
	}, nil
}

// epsilonGreedy randomly selects arms with a probability of ε. The rest of
// the time, epsilonGreedy selects the currently best known arm.
type epsilonGreedy struct {
	counts  []int      // number of pulls. len(counts) == arms.
	values  []float64  // running average reward per arm. len(values) == arms.
	arms    int        // number of arms present in this bandit
	epsilon float64    // epsilon value for this bandit
	rand    *rand.Rand // seeded random number generator
}

// SelectArm returns 1 indexed arm to be tried next.
func (e *epsilonGreedy) SelectArm() int {
	arm := 0
	if z := e.rand.Float64(); z > e.epsilon {
		imax, max := []int{}, 0.0
		for i, value := range e.values {
			if value > max {
				max = value
				imax = []int{i}
			} else if value == max {
				imax = append(imax, i)
			}
		}

		// best arm. randomly pick because there may be equally best arms.
		arm = imax[e.rand.Intn(len(imax))]
	} else {
		// random arm
		arm = e.rand.Intn(e.arms)
	}

	e.counts[arm]++
	return arm + 1
}

// Update the running average, where arm is the 1 indexed arm
func (e *epsilonGreedy) Update(arm int, reward float64) {
	arm--
	e.counts[arm]++
	count := e.counts[arm]
	e.values[arm] = ((e.values[arm] * float64(count-1)) + reward) / float64(count)
}

// Version returns information on this bandit
func (e *epsilonGreedy) Version() string {
	return fmt.Sprintf("EpsilonGreedy(epsilon=%.2f)", e.epsilon)
}

// Reset returns the bandit to it's newly constructed state
func (e *epsilonGreedy) Reset() {
	e.counts = make([]int, e.arms)
	e.values = make([]float64, e.arms)
	e.rand = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// NewSoftmax constructs a softmax bandit. Softmax explores arms in proportion
// to their estimated values.
func NewSoftmax(arms int, τ float64) (Bandit, error) {
	if !(τ >= 0.0) {
		return &softmax{}, fmt.Errorf("τ not in [0, ∞]")
	}

	return &softmax{
		counts: make([]int, arms),
		values: make([]float64, arms),
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
		arms:   arms,
		tau:    τ,
	}, nil
}

// softmax holds counts values and temperature τ
type softmax struct {
	counts []int      // number of pulls. len(counts) == arms.
	values []float64  // running average reward per arm. len(values) == arms.
	arms   int        // number of arms present in this bandit
	tau    float64    // tau value for this bandit
	rand   *rand.Rand // seeded random number generator
}

// SelectArm returns 1 indexed arm to be tried next.
func (s *softmax) SelectArm() int {
	normalizer := 0.0
	for _, value := range s.values {
		normalizer += math.Exp(value / s.tau)
	}

	cumulativeProb := 0.0
	draw := len(s.values) - 1
	z := s.rand.Float64()
	for i, value := range s.values {
		cumulativeProb = cumulativeProb + math.Exp(value/s.tau)/normalizer
		if cumulativeProb > z {
			draw = i
			break
		}
	}

	return draw + 1
}

// Update the running average, where arm is the 1 indexed arm.
func (s *softmax) Update(arm int, reward float64) {
	arm--
	s.counts[arm]++
	count := s.counts[arm]
	s.values[arm] = ((s.values[arm] * float64(count-1)) + reward) / float64(count)
}

// Version returns information on this bandit
func (s *softmax) Version() string {
	return fmt.Sprintf("Softmax(tau=%.2f)", s.tau)
}

// Reset returns the bandit to it's newly constructed state
func (s *softmax) Reset() {
	s.counts = make([]int, s.arms)
	s.values = make([]float64, s.arms)
	s.rand = rand.New(rand.NewSource(time.Now().UnixNano()))
}
