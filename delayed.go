package bandit

import "fmt"

// NewDelayedBandit wraps the given bandit.
func NewDelayedBandit(b *Bandit, updates <-chan Counters) (Bandit, error) {
	bandit := delayedBandit{
		bandit:  b,
		updates: updates,
	}

	go func() {
		for counters := range updates {
			(*b).Reset(&counters)
		}
	}()

	return &bandit, nil
}

// delayedBandit wraps a bandit. Internal counters are stored at the
// configured source file, which is pooled at `poll` interval. The retrieved
// Snapshot replaces the bandit's internal counters.
type delayedBandit struct {
	Counters
	updates <-chan Counters
	bandit  *Bandit
}

// SelectArm delegates to the wrapped bandit
func (b *delayedBandit) SelectArm() int {
	return (*b.bandit).SelectArm()
}

// Version gives information about delayed bandit + the wrapped bandit.
func (b *delayedBandit) Version() string {
	return fmt.Sprintf("Delayed(%s)", (*b.bandit).Version())
}

// DelayedUpdate updates the internal counters of a bandit with the provided
// counters.
func (b *delayedBandit) Reset(c *Counters) error {
	b.Lock()
	defer b.Unlock()
	return (*b.bandit).Reset(c)
}

// Update is a NOP. Delayed bandit is updated with Reset(counter) instead
func (b *delayedBandit) Update(arm int, reward float64) {}
