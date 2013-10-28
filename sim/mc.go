// Copyright 2013 SoundCloud, Rany Keddo. All rights reserved.  Use of this
// source code is governed by a license that can be found in the LICENSE file.

package sim

import (
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

// Arm simulates a single bandit arm pull with every execution. Returns {0,1}.
type Arm func() float64

// Gaussian returns an Arm function such that a ~ N(x|μ,σ)
func Gaussian(μ, σ float64) Arm {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return func() float64 {
		return r.NormFloat64()*σ + μ
	}
}

// Constant returns an Arm function such that a ~ c
func Constant(c float64) Arm {
	return func() float64 {
		return c
	}
}

// Bernoulli returns an Arm function such that a ~ Bern(x|μ)
func Bernoulli(μ float64) Arm {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return func() float64 {
		res := 0.0
		if r.Float64() <= μ {
			res = 1.0
		}

		return res
	}
}

// MonteCarlo runs a monte carlo experiment with the given bandit and arms.
func MonteCarlo(sims, trials int, arms []Arm, b Bandit) (Simulation, error) {
	s := Simulation{
		Sims:       sims,
		Trials:     trials,
		Sim:        make([]int, sims*trials),
		Trial:      make([]int, sims*trials),
		Selected:   make([]int, sims*trials),
		Reward:     make([]float64, sims*trials),
		Cumulative: make([]float64, sims*trials),
	}
	for sim := 0; sim < sims; sim++ {
		s.Description = b.Version()
		b.Reset()

		for trial := 0; trial < trials; trial++ {
			selected := b.SelectArm()
			reward := arms[selected-1]()
			b.Update(selected, reward)

			// record this trial into column i
			i := sim*trials + trial
			s.Sim[i] = sim + 1
			s.Trial[i] = trial + 1
			s.Selected[i] = selected
			s.Reward[i] = reward
			if trial == 0 {
				s.Cumulative[i] = 0.0
			} else {
				s.Cumulative[i] = s.Cumulative[i-1] + reward
			}
		}
	}

	return s, nil
}

// Simulation is a matrix of simulation results. Columns represent individual
// trial results that grow to the right with each trial
type Simulation struct {
	Sims        int
	Trials      int
	Description string
	Sim         []int
	Trial       []int
	Selected    []int
	Reward      []float64
	Cumulative  []float64
}

// Summary summarizes a Simulation and returns corresponding plot points.
type Summary func(s *Simulation) []float64

// Accuracy returns the proportion of times the best arm was pulled at each
// trial point. Takes a slice of best arms since n arms may be equally good.
func Accuracy(bestArms []int) Summary {
	return func(s *Simulation) []float64 {
		t := make([]float64, s.Trials)
		for trial := 0; trial < s.Trials; trial++ {
			correct := 0
			for sim := 0; sim < s.Sims; sim++ {
				i := sim*s.Trials + trial
				if s.Trial[i] != trial+1 {
					panic("impossible trial access")
				}

				for _, best := range bestArms {
					if s.Selected[i] == best {
						correct = correct + 1
					}
				}
			}

			t[trial] = float64(correct) / float64(s.Sims)
		}

		return t
	}
}

// Performance returns an array of mean rewards at each trial point.
// Averaged over sims
func Performance(s *Simulation) []float64 {
	t := make([]float64, s.Trials)
	for trial := 0; trial < s.Trials; trial++ {
		accum, count := 0.0, 0
		for sim := 0; sim < s.Sims; sim++ {
			i := sim*s.Trials + trial
			if s.Trial[i] != trial+1 {
				panic("impossible trial access")
			}

			accum = accum + s.Reward[i]
			count = count + 1
		}

		t[trial] = accum / float64(count)
	}

	return t
}

// Cumulative performance returns an array of mean rewards at each trial
// point.  Averaged over sims
func Cumulative(s *Simulation) []float64 {
	t := make([]float64, s.Trials)
	for trial := 0; trial < s.Trials; trial++ {
		accum, count := 0.0, 0
		for sim := 0; sim < s.Sims; sim++ {
			i := sim*s.Trials + trial
			if s.Trial[i] != trial+1 {
				panic("impossible trial access")
			}

			accum = accum + s.Cumulative[i]
			count = count + 1
		}

		t[trial] = accum / float64(count)
	}

	return t
}
