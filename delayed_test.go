package bandit

import "testing"

func TestDelayedBandit(t *testing.T) {
	τ := 0.1
	sims := 5000
	trials := 300
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

	updates := make(chan Counters)
	d, err := NewDelayedBandit(&b, updates)
	if err != nil {
		t.Fatalf(err.Error())
	}

	_, err = MonteCarlo(sims, trials, arms, d)
	if err != nil {
		t.Fatalf(err.Error())
	}

	// TODO: test accuracy, perforamce and cumulative like in the other tests.
	// Need to grab snapshot side channel and pass it back to the delayed bandit
	// during test for this to work.
}
