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
	"sync"
	"time"
)

// Bandit can select arm or update information
type Bandit interface {
	SelectArm() int
	Update(arm int, reward float64)
	Version() string
	Reset(*Counters) error
	Init()
}

// epsilonGreedy randomly selects arms with a probability of ε. The rest of
// the time, epsilonGreedy selects the currently best known arm.
type epsilonGreedy struct {
	Counters
	epsilon float64 // epsilon value for this bandit
}

// NewEpsilonGreedy constructs an epsilon greedy bandit.
func NewEpsilonGreedy(arms int, epsilon float64) (Bandit, error) {
	if !(epsilon >= 0 && epsilon <= 1) {
		return &epsilonGreedy{}, fmt.Errorf("epsilon not in [0, 1]")
	}

	return &epsilonGreedy{
		Counters: NewCounters(arms),
		epsilon:  epsilon,
	}, nil
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

// Version returns information on this bandit
func (e *epsilonGreedy) Version() string {
	return fmt.Sprintf("EpsilonGreedy(epsilon=%.2f)", e.epsilon)
}

// softmax selects proportially to success
type softmax struct {
	Counters
	tau float64 // tau value for this bandit
}

// NewSoftmax constructs a softmax bandit. Softmax explores arms in proportion
// to their estimated values.
func NewSoftmax(arms int, τ float64) (Bandit, error) {
	if !(τ >= 0.0) {
		return &softmax{}, fmt.Errorf("τ not in [0, ∞]")
	}

	return &softmax{
		Counters: NewCounters(arms),
		tau:      τ,
	}, nil
}

type categorialDistribution struct {
	individualProbs []float64 // (unnormalized) probability that a given observation will be in category i
	normalizer      float64
}

// identify bin of cumulative sum of bins which contains value
// TODO(cs): error handling if id == len
func getBinId(x float64, bins []float64) int {
	id := 0
	for cum := bins[0]; cum < x; cum += bins[id] {
		id++
		if id == len(bins) { // bins do not contain value x
			return -1
		}
	}
	return id
}

// Categorically distributed random numbers
func CatRand(dist categorialDistribution) int {
	return getBinId(rand.Float64()*dist.normalizer, dist.individualProbs) + 1
}

// TODO(cs): extract probability distribution stuff into separate file (catRand, Max, Sum,...)
func Max(values []float64) (max float64, imax []int) {
	imax, max = []int{}, -math.MaxFloat64
	for i, value := range values {
		if value > max {
			max = value
			imax = []int{i}
		} else if value == max {
			imax = append(imax, i)
		}
	}
	return max, imax
}

// calculate soft max probs, i.e., p(x_i) = exp(score_i/tau)/(sum_j exp(score_j/tau))
func calcSoftMax(scores []float64, tau float64) categorialDistribution {
	probs := categorialDistribution{individualProbs: make([]float64, len(scores))}
	max, _ := Max(scores)
	// calculate exponentials and normalizer (numerical stable)
	for i, value := range scores {
		probs.individualProbs[i] = math.Exp((value - max) / tau) // subtract maximum to avoid numerical problems
		probs.normalizer += probs.individualProbs[i]
	}
	return probs
}

// TODO(cs): s.values have to be normalized (mean and variance should be realistic)
// SelectArm returns 1 indexed arm to be tried next.
func (s *softmax) SelectArm() int {
	dist := calcSoftMax(s.values, s.tau)
	draw := CatRand(dist)
	s.counts[draw-1]++
	return draw
}

// Version returns information on this bandit
func (s *softmax) Version() string {
	return fmt.Sprintf("Softmax(tau=%.2f)", s.tau)
}

// NewUCB1 returns a UCB1 bandit
func NewUCB1(arms int) Bandit {
	return &uCB1{
		Counters: NewCounters(arms),
	}
}

// uCB1
type uCB1 struct {
	Counters
}

// SelectArm returns 1 indexed arm to be tried next.
func (u *uCB1) SelectArm() int {
	for i, count := range u.counts {
		if count == 0 {
			u.counts[i]++
			return i + 1
		}
	}

	var totalCounts int
	for _, count := range u.counts {
		totalCounts += count
	}

	ucbValues := make([]float64, u.arms)
	for i := 0; i < u.arms; i++ {
		bonus := math.Sqrt((2 * math.Log(float64(totalCounts))) / float64(u.counts[i]))
		ucbValues[i] = u.values[i] + bonus
	}

	var arm int
	var max float64
	for i, val := range ucbValues {
		if max < val {
			arm = i
			max = val
		}
	}
	u.counts[arm]++
	return arm + 1
}

// Version returns information on this bandit
func (u *uCB1) Version() string {
	return fmt.Sprintf("UCB1")
}

// Counters maintain internal bandit state
type Counters struct {
	sync.Mutex

	arms   int        // number of arms present in this bandit
	counts []int      // number of pulls. len(counts) == arms.
	rand   *rand.Rand // seeded random number generator
	values []float64  // running average reward per arm. len(values) == arms.
}

// NewCounters constructs counters for given arms
func NewCounters(arms int) Counters {
	return Counters{
		arms:   arms,
		counts: make([]int, arms),
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
		values: make([]float64, arms),
	}
}

// Update the running average, where arm is the 1 indexed arm
func (c *Counters) Update(arm int, reward float64) {
	c.Lock()
	defer c.Unlock()

	arm--
	count := c.counts[arm]
	c.values[arm] = ((c.values[arm] * float64(count-1)) + reward) / float64(count)
}

// Reset the bandit to a new counter state. Pass in Counters{} to reset to
// initial state.
func (c *Counters) Reset(snapshot *Counters) error {
	if c.arms != snapshot.arms {
		return fmt.Errorf("cannot %d arms with %d arms", c.arms, snapshot.arms)
	}

	if snapshot.arms == 0 {
		return fmt.Errorf("need at least 1 arm")
	}

	c.Lock()
	defer c.Unlock()

	c.counts = snapshot.counts
	c.rand = snapshot.rand
	c.values = snapshot.values

	return nil
}

// Init reset the bandit to initial state.
func (c *Counters) Init() {
	c.Lock()
	defer c.Unlock()

	// TODO(cs): error handling
	// TODO(cs): will mostly fail since number of arms are typically non-zero
	c.Reset(&Counters{})
}
