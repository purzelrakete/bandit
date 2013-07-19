package bandit

import (
	"fmt"
	"math/rand"
	"time"
)

// Bandit can select arm or update information
type Bandit interface {
	SelectArm() int
	Update(arm int, reward float64)
}

// EpsilonGreedyNew constructs an epsilon greedy bandit.
func EpsilonGreedyNew(arms int, epsilon float64) (Bandit, error) {
	if !(epsilon >= 0 && epsilon <= 1) {
		return epsilonGreedy{}, fmt.Errorf("epsilon not in [0, 1]")
	}

	return epsilonGreedy{
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
	counts  []int
	values  []float64
	epsilon float64
	arms    int
	rand    *rand.Rand
}

// SelectArm according to EpsilonGreedy strategy
func (e epsilonGreedy) SelectArm() int {
	arm := 0
	if e.rand.Float64() > e.epsilon {
		// best arm
		for i := range e.values {
			if e.values[i] > e.values[arm] {
				arm = i
			}
		}
	} else {
		// random arm
		arm = e.rand.Intn(e.arms)
	}

	e.counts[arm] = e.counts[arm] + 1
	return arm + 1
}

// Update the running average
func (e epsilonGreedy) Update(arm int, reward float64) {
	arm = arm - 1
	e.counts[arm] = e.counts[arm] + 1
	count := e.counts[arm]
	e.values[arm] = ((e.values[arm] * float64(count-1)) + reward) / float64(count)
}
