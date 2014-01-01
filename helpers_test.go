package bandit

// NewSimulatedDelayedStrategy simulates delayed strategy by flushing counters to
// the underlying strategy after `flush` number of updates.
func NewSimulatedDelayedStrategy(b Strategy, arms, flush int) Strategy {
	return &simulatedDelayedStrategy{
		limit:   flush,
		updates: flush,
		delayedStrategy: delayedStrategy{
			strategy: b,
			Counters: NewCounters(arms),
		},
	}
}

// simulatedDelayedStrategy is used for testing but also simulation and
// plotting. It simulates delayed strategy by flushing counters to the
// underlying strategy after `limit` number of updates.
type simulatedDelayedStrategy struct {
	delayedStrategy
	limit   int // #updates to wait before flushing Counters to underlying strategy
	updates int // #updates since last flush
}

// Update flushes counters to the underlying strategy every n updates. This is
// approximately the behaviour seen by a delayed strategy in production.
func (b *simulatedDelayedStrategy) Update(arm int, reward float64) {
	b.Lock()
	defer b.Unlock()

	arm--
	b.counts[arm]++
	count := b.counts[arm]
	b.values[arm] = ((b.values[arm] * float64(count-1)) + reward) / float64(count)

	b.updates++
	if b.updates >= b.limit {
		b.strategy.Init(&b.Counters)
		b.updates = 0
	}
}
