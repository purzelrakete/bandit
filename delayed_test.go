package bandit

import (
	"math"
	"testing"
)

func TestDelayedBandit(t *testing.T) {
	τ := 0.1
	sims := 5000
	trials := 300
	bestArmIndex := 4 // Bernoulli(bestArm)
	bestArm := 0.8
	arms := []Arm{
		Bernoulli(0.1),
		Bernoulli(0.3),
		Bernoulli(0.2),
		Bernoulli(0.8),
	}

	b, err := NewSoftmax(len(arms), τ)
	if err != nil {
		t.Fatalf(err.Error())
	}

	d := testDelayedBandit{
		limit:   100,
		updates: 100,
		delayedBandit: delayedBandit{
			bandit:   &b,
			Counters: NewCounters(len(arms)),
		},
	}

	if err != nil {
		t.Fatalf(err.Error())
	}

	sim, err := MonteCarlo(sims, trials, arms, &d)
	if err != nil {
		t.Fatalf(err.Error())
	}

	accuracies := Accuracy([]int{bestArmIndex})(&sim)
	if got := accuracies[len(accuracies)-1]; got < 0.9 {
		t.Fatalf("accuracy is only %f. %d sims, %d trials", got, sims, trials)
	}

	performances := Performance(&sim)
	if got := performances[len(performances)-1]; math.Abs(bestArm-got) > 0.1 {
		t.Fatalf("performance converge to %f. is %f", bestArm, got)
	}

	expectedCumulative := 200.0
	cumulatives := Cumulative(&sim)
	if got := cumulatives[len(cumulatives)-1]; got < expectedCumulative {
		t.Fatalf("cumulative performance should be > %f. is %f", expectedCumulative, got)
	}
}

// testing bandit
type testDelayedBandit struct {
	delayedBandit
	limit   int // #updates to wait before flushing Counters to underlying bandit
	updates int // #updates since last flush
}

// Update flushes counters to the underlying bandit every n updates. This is
// approximately the behaviour seen by a delayed bandit in production.
func (b *testDelayedBandit) Update(arm int, reward float64) {
	b.Lock()
	defer b.Unlock()

	arm--
	b.counts[arm]++
	count := b.counts[arm]
	b.values[arm] = ((b.values[arm] * float64(count-1)) + reward) / float64(count)

	b.updates++
	if b.updates >= b.limit {
		(*b.bandit).Reset(&b.Counters)
		b.updates = 0
	}
}
