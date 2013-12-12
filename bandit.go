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
	bmath "github.com/purzelrakete/bandit/math"
	"log"
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
	Init(*Counters) error
	Reset()
}

// NewBandit returns an initialized bandit given a string name such as
// 'softmax'.
func NewBandit(arms int, name string, params []float64) (Bandit, error) {
	switch name {
	case "epsilonGreedy":
		if len(params) != 1 {
			return &epsilonGreedy{}, fmt.Errorf("missing ε")
		}

		return NewEpsilonGreedy(arms, params[0])
	case "uniform":
		if len(params) != 0 {
			return &epsilonGreedy{}, fmt.Errorf("uniform has no parameters")
		}

		return NewEpsilonGreedy(arms, 0)
	case "softmax":
		if len(params) != 1 {
			return &softmax{}, fmt.Errorf("missing τ")
		}

		return NewSoftmax(arms, params[0])
	case "ucb1":
		if len(params) != 0 {
			return &softmax{}, fmt.Errorf("UCB1 has no parameters")
		}

		return NewUCB1(arms), nil
	case "thompson":
		if len(params) != 1 {
			return &thompson{}, fmt.Errorf("missing α")
		}

		return NewThompson(arms, params[0])
	}

	return &epsilonGreedy{}, fmt.Errorf("'%s' unknown bandit", name)
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
		_, imax := max(e.values)
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
		return &softmax{}, fmt.Errorf("τ not in [0, ∞)")
	}

	return &softmax{
		Counters: NewCounters(arms),
		tau:      τ,
	}, nil
}

// SelectArm returns 1 indexed arm to be tried next.
func (s *softmax) SelectArm() int {
	max, _ := max(s.values)

	normalizer := 0.0
	for _, value := range s.values {
		normalizer += math.Exp((value - max) / s.tau)
	}

	if math.IsInf(normalizer, 0) {
		panic("normalizer in softmax too large")
	}

	cumulativeProb := 0.0
	draw := len(s.values) - 1
	z := s.rand.Float64()
	for i, value := range s.values {
		cumulativeProb = cumulativeProb + math.Exp((value-max)/s.tau)/normalizer
		if cumulativeProb > z {
			draw = i
			break
		}
	}
	s.counts[draw]++
	return draw + 1
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

	_, imax := max(ucbValues)
	// best arm. randomly pick because there may be equally best arms.
	arm := imax[u.rand.Intn(len(imax))]

	u.counts[arm]++
	return arm + 1
}

// Version returns information on this bandit
func (u *uCB1) Version() string {
	return fmt.Sprintf("UCB1")
}

// NewDelayedBandit wraps a bandit and updates internal counters from
// a snapshot at `poll` interval.
func NewDelayedBandit(b Bandit, o Opener, poll time.Duration) (Bandit, error) {
	// fail once
	if _, err := GetSnapshot(o); err != nil {
		return &delayedBandit{}, fmt.Errorf("could not get snapshot: %s", err.Error())
	}

	c := make(chan Counters)
	go func() {
		t := time.NewTicker(poll)
		for _ = range t.C {
			counters, err := GetSnapshot(o)
			if err != nil {
				log.Printf("BanditError: could not get snapshot: %s", err.Error())
			}

			c <- counters
		}
	}()

	bandit := delayedBandit{
		bandit:  b,
		updates: c,
	}

	go func() {
		for counters := range c {
			b.Init(&counters)
		}
	}()

	return &bandit, nil
}

// delayedBandit wraps a bandit. Internal counters are stored at the
// configured source file, which is pooled at `poll` interval. The retrieved
// Snapshot replaces the bandit's internal counters.
type delayedBandit struct {
	Counters
	updates chan Counters
	bandit  Bandit
}

// SelectArm delegates to the wrapped bandit
func (b *delayedBandit) SelectArm() int {
	return b.bandit.SelectArm()
}

// Version gives information about delayed bandit + the wrapped bandit.
func (b *delayedBandit) Version() string {
	return fmt.Sprintf("Delayed(%s)", b.bandit.Version())
}

// DelayedUpdate updates the internal counters of a bandit with the provided
// counters.
func (b *delayedBandit) Init(c *Counters) error {
	b.Lock()
	defer b.Unlock()
	return b.bandit.Init(c)
}

// Update is a NOP. Delayed bandit is updated with Reset(counter) instead
func (b *delayedBandit) Update(arm int, reward float64) {}

// Thompson sampling (for Bernoulli bandits) explores arms by sampling
// according to the probability that it maximizes the expected reward.
type thompson struct {
	Counters
	betaRnd func(a, b float64) float64
	alpha   float64 // strength of prior distribution for each bandit (beta with homogeneous prior)
}

// NewThompson constructs a thompson sampling strategy.
func NewThompson(arms int, α float64) (Bandit, error) {
	if !(α > 0.0) {
		return &thompson{}, fmt.Errorf("α not in (0, ∞]")
	}

	return &thompson{
		Counters: NewCounters(arms),
		alpha:    α,
		betaRnd:  bmath.BetaRnd(),
	}, nil
}

// SelectArm returns 1 indexed arm to be tried next.
func (t *thompson) SelectArm() int {
	var thetas = make([]float64, t.arms)
	for i := 0; i < t.arms; i++ {
		si := t.values[i] * float64(t.counts[i])
		fi := float64(t.counts[i]) - si
		thetas[i] = t.betaRnd(si+t.alpha, fi+t.alpha)
	}

	_, imax := max(thetas)
	// best arm. randomly pick because there may be equally best arms.
	arm := imax[t.rand.Intn(len(imax))]

	t.counts[arm]++
	return arm + 1
}

// Version returns information on this bandit
func (t *thompson) Version() string {
	return fmt.Sprintf("Thompson(alpha=%.2f)", t.alpha)
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

// Init the bandit to a new counter state.
func (c *Counters) Init(snapshot *Counters) error {
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

// Reset the bandit to initial state.
func (c *Counters) Reset() {
	c.counts = make([]int, c.arms)
	c.values = make([]float64, c.arms)
}

// helpers

// max returns maximal value and its indices of a slice
func max(array []float64) (float64, []int) {
	max, imax := -math.MaxFloat64, []int{}
	for idx, value := range array {
		if max < value {
			imax = []int{idx}
			max = value
		} else if value == max {
			imax = append(imax, idx)
		}
	}
	return max, imax
}
