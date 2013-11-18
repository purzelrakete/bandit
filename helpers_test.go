package bandit

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
		b.bandit.Init(&b.Counters)
		b.updates = 0
	}
}
