package bandit

import (
	"math"
	"math/rand"
	"time"
)

// Arm simulates a single bandit arm pull with every execution. Returns {0,1}.
type Arm func() float64

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

// BanditNew is a curried constructor function.
type BanditNew func() (Bandit, error)

// MonteCarlo runs a monte carlo experiment with the given bandit and arms.
func MonteCarlo(sims, horizon int, bandit BanditNew, arms []Arm) (Sim, error) {
	s := Sim{
		Selected:   make([]int, sims*horizon),
		Reward:     make([]float64, sims*horizon),
		Cumulative: make([]float64, sims*horizon),
		Sim:        make([]int, sims*horizon),
		Trial:      make([]int, sims*horizon),
	}

	for sim := 0; sim < sims; sim++ {
		b, err := bandit()
		if err != nil {
			return Sim{}, err
		}

		for trial := 0; trial < horizon; trial++ {
			selected := b.SelectArm()
			reward := arms[selected-1]()
			b.Update(selected, reward)

			// record this trial into column i
			i := sim*horizon + trial
			s.Selected[i] = selected
			s.Reward[i] = reward
			s.Cumulative[i] = s.Cumulative[int(math.Max(float64(i-1), 0.0))] + reward
			s.Sim[i] = sim + 1
			s.Trial[i] = trial + 1
		}
	}

	return s, nil
}

// Sim is a matrix of simulation results. Columns represent individual trial
// results that grow to the right with each trial simulation
type Sim struct {
	Selected   []int
	Reward     []float64
	Cumulative []float64
	Sim        []int
	Trial      []int
}
