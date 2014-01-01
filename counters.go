package bandit

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// NewCounters constructs counters for given arms
func NewCounters(arms int) Counters {
	return Counters{
		arms:   arms,
		counts: make([]int, arms),
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
		values: make([]float64, arms),
	}
}

// Counters maintain internal strategy state
type Counters struct {
	sync.Mutex

	arms   int        // number of arms present in this strategy
	counts []int      // number of pulls. len(counts) == arms.
	rand   *rand.Rand // seeded random number generator
	values []float64  // running average reward per arm. len(values) == arms.
}

// Update the running average, where arm is the 1 indexed arm
func (c *Counters) Update(arm int, reward float64) {
	c.Lock()
	defer c.Unlock()

	arm--
	count := c.counts[arm]
	c.values[arm] = ((c.values[arm] * float64(count-1)) + reward) / float64(count)
}

// Init the strategy to a new counter state.
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

// Reset the strategy to initial state.
func (c *Counters) Reset() {
	c.counts = make([]int, c.arms)
	c.values = make([]float64, c.arms)
}
