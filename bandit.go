// Copyright 2013 SoundCloud, Rany Keddo. All rights reserved.  Use of this
// source code is governed by a license that can be found in the LICENSE file.

// Package bandit implements a multiarmed strategy. Runs A/B tests while
// controlling the tradeoff between exploring new arms and exploiting the
// currently best arm.
//
// The project is based on John Myles White's 'Strategy Algorithms for Website
// Optimization'.
package bandit

import (
	"fmt"
	bmath "github.com/purzelrakete/bandit/math"
	"log"
	"math"
	"time"
)

// Strategy can select arm or update information
type Strategy interface {
	SelectArm() int
	Update(arm int, reward float64)
	Init(*Counters) error
	Reset()
}

// New returns an initialized stragtegy given a name like 'softmax'.
func New(arms int, name string, params []float64) (Strategy, error) {
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

		return NewEpsilonGreedy(arms, 1)
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

	return &epsilonGreedy{}, fmt.Errorf("'%s' unknown strategy", name)
}

// NewEpsilonGreedy constructs an epsilon greedy strategy.
func NewEpsilonGreedy(arms int, epsilon float64) (Strategy, error) {
	if !(epsilon >= 0 && epsilon <= 1) {
		return &epsilonGreedy{}, fmt.Errorf("epsilon not in [0, 1]")
	}

	return &epsilonGreedy{
		Counters: NewCounters(arms),
		epsilon:  epsilon,
	}, nil
}

// epsilonGreedy randomly selects arms with a probability of ε. The rest of
// the time, epsilonGreedy selects the currently best known arm.
type epsilonGreedy struct {
	Counters
	epsilon float64 // epsilon value for this strategy
}

// SelectArm returns 1 indexed arm to be tried next.
func (e *epsilonGreedy) SelectArm() int {
	arm := 0
	if z := e.rand.Float64(); z > e.epsilon {
		_, imax := bmath.Max(e.values)
		// best arm. randomly pick because there may be equally best arms.
		arm = imax[e.rand.Intn(len(imax))]
	} else {
		// random arm
		arm = e.rand.Intn(e.arms)
	}

	e.counts[arm]++
	return arm + 1
}

// String returns information on this strategy
func (e *epsilonGreedy) String() string {
	return fmt.Sprintf("EpsilonGreedy(epsilon=%.2f)", e.epsilon)
}

// NewSoftmax constructs a softmax strategy. Softmax explores arms in proportion
// to their estimated values.
func NewSoftmax(arms int, τ float64) (Strategy, error) {
	if !(τ >= 0.0) {
		return &softmax{}, fmt.Errorf("τ not in [0, ∞)")
	}

	return &softmax{
		Counters: NewCounters(arms),
		tau:      τ,
	}, nil
}

// softmax selects proportially to success
type softmax struct {
	Counters
	tau float64 // tau value for this Strategy
}

// SelectArm returns 1 indexed arm to be tried next.
func (s *softmax) SelectArm() int {
	max, _ := bmath.Max(s.values)

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

// String returns information on this Strategy
func (s *softmax) String() string {
	return fmt.Sprintf("Softmax(tau=%.2f)", s.tau)
}

// NewUCB1 returns a UCB1 Strategy
func NewUCB1(arms int) Strategy {
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

	_, imax := bmath.Max(ucbValues)
	// best arm. randomly pick because there may be equally best arms.
	arm := imax[u.rand.Intn(len(imax))]

	u.counts[arm]++
	return arm + 1
}

// String returns information on this Strategy
func (u *uCB1) String() string {
	return fmt.Sprintf("UCB1")
}

// NewDelayed wraps a strategy and updates internal counters from a snapshot at
// `poll` interval.
func NewDelayed(s Strategy, o Opener, poll time.Duration) (Strategy, error) {
	// fail once
	if _, err := GetSnapshot(o); err != nil {
		return &delayedStrategy{}, fmt.Errorf("could not get snapshot: %s", err.Error())
	}

	c := make(chan Counters)
	go func() {
		t := time.NewTicker(poll)
		for _ = range t.C {
			counters, err := GetSnapshot(o)
			if err != nil {
				log.Printf("Error: could not get snapshot: %s", err.Error())
			}

			c <- counters
		}
	}()

	strategy := delayedStrategy{
		strategy: s,
		updates:  c,
	}

	go func() {
		for counters := range c {
			s.Init(&counters)
		}
	}()

	return &strategy, nil
}

// delayedStrategy wraps a strategy. Internal counters are stored at the
// configured source file, which is pooled at `poll` interval. The retrieved
// Snapshot replaces the strategy's internal counters.
type delayedStrategy struct {
	Counters
	updates  chan Counters
	strategy Strategy
}

// SelectArm delegates to the wrapped strategy
func (b *delayedStrategy) SelectArm() int {
	return b.strategy.SelectArm()
}

// String gives information about delayed strategy + the wrapped strategy.
func (b *delayedStrategy) String() string {
	return fmt.Sprintf("Delayed(%b)", b.strategy)
}

// DelayedUpdate updates the internal counters of a strategy with the provided
// counters.
func (b *delayedStrategy) Init(c *Counters) error {
	b.Lock()
	defer b.Unlock()
	return b.strategy.Init(c)
}

// Update is a NOP. Delayed strategy is updated with Reset(counter) instead
func (b *delayedStrategy) Update(arm int, reward float64) {}

// NewThompson constructs a thompson sampling strategy.
func NewThompson(arms int, α float64) (Strategy, error) {
	if !(α > 0.0) {
		return &thompson{}, fmt.Errorf("α not in (0, ∞]")
	}

	return &thompson{
		Counters: NewCounters(arms),
		alpha:    α,
		betaRand: bmath.NewBetaRand(time.Now().UnixNano()),
	}, nil
}

// Thompson sampling (for Bernoulli strategys) explores arms by sampling
// according to the probability that it maximizes the expected reward.
type thompson struct {
	Counters
	betaRand *bmath.BetaRand
	alpha    float64 // strength of prior distributionr. beta with homogeneous prior
}

// SelectArm returns 1 indexed arm to be tried next.
func (t *thompson) SelectArm() int {
	var thetas = make([]float64, t.arms)
	for i := 0; i < t.arms; i++ {
		si := t.values[i] * float64(t.counts[i])
		fi := float64(t.counts[i]) - si
		thetas[i] = t.betaRand.NextBeta(si+t.alpha, fi+t.alpha)
	}

	_, imax := bmath.Max(thetas)
	// best arm. randomly pick because there may be equally best arms.
	arm := imax[t.rand.Intn(len(imax))]

	t.counts[arm]++
	return arm + 1
}

// String returns information on this strategy
func (t *thompson) String() string {
	return fmt.Sprintf("Thompson(alpha=%.2f)", t.alpha)
}
