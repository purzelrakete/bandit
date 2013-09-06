package bandit

import (
	"fmt"
	"log"
	"time"
)

// NewDelayedBandit wraps the given bandit.
func NewDelayedBandit(b Bandit, updates chan Counters) (Bandit, error) {
	bandit := delayedBandit{
		bandit:  b,
		updates: updates,
	}

	go func() {
		for counters := range updates {
			b.Reset(&counters)
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
func (b *delayedBandit) Reset(c *Counters) error {
	b.Lock()
	defer b.Unlock()
	return b.bandit.Reset(c)
}

// Update is a NOP. Delayed bandit is updated with Reset(counter) instead
func (b *delayedBandit) Update(arm int, reward float64) {}

// NewDelayedTrials constructs a (bandit, experiment) tuples
func NewDelayedTrials(experiment, snapshot string, poll time.Duration) (Trials, error) {
	c := make(chan Counters)
	go func() {
		t := time.NewTicker(poll)
		for _ = range t.C {
			counters, err := GetSnapshot(snapshot)
			if err != nil {
				log.Fatalf("could not get snapshot: %s", err.Error())
			}

			c <- counters
		}
	}()

	trials, err := NewTrials(experiment, func(arms int) (Bandit, error) {
		b, err := NewSoftmax(arms, 0.1)
		if err != nil {
			return b, err
		}

		return NewDelayedBandit(b, c)
	})

	if err != nil {
		log.Fatalf("could not construct experiments: %s", err.Error())
	}

	return trials, nil
}

// NewSimulatedDelayedBandit simulates delayed bandit by flushing counters to
// the underlying bandit after `flush` number of updates.
func NewSimulatedDelayedBandit(b Bandit, arms, flush int) Bandit {
	return &simulatedDelayedBandit{
		limit:   flush,
		updates: flush,
		delayedBandit: delayedBandit{
			bandit:   b,
			Counters: NewCounters(arms),
		},
	}
}

// simulatedDelayedBandit is used for testing but also simulation and
// plotting. It simulates delayed bandit by flushing counters to the
// underlying bandit after `limit` number of updates.
type simulatedDelayedBandit struct {
	delayedBandit
	limit   int // #updates to wait before flushing Counters to underlying bandit
	updates int // #updates since last flush
}

// Update flushes counters to the underlying bandit every n updates. This is
// approximately the behaviour seen by a delayed bandit in production.
func (b *simulatedDelayedBandit) Update(arm int, reward float64) {
	b.Lock()
	defer b.Unlock()

	arm--
	b.counts[arm]++
	count := b.counts[arm]
	b.values[arm] = ((b.values[arm] * float64(count-1)) + reward) / float64(count)

	b.updates++
	if b.updates >= b.limit {
		b.bandit.Reset(&b.Counters)
		b.updates = 0
	}
}
