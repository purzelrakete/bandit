package bandit

import (
	"testing"
)

func TestEpsilonGreedy(t *testing.T) {
	ε := 0.1
	sims := 2
	trials := 5
	arms := []Arm{
		Bernoulli(0.1),
		Bernoulli(0.3),
		Bernoulli(0.2),
		Bernoulli(0.8),
	}

	bandit, err := EpsilonGreedyNew(len(arms), ε)
	if err != nil {
		t.Fatalf(err.Error())
	}

	sim, err := MonteCarlo(sims, trials, arms, bandit)
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := sims * trials
	if got := len(sim.Selected); got != expected {
		t.Fatalf("incorrect number of trials: %d", got)
	}
}

func TestSoftmax(t *testing.T) {
	τ := 0.1
	sims := 2
	trials := 5
	arms := []Arm{
		Bernoulli(0.1),
		Bernoulli(0.3),
		Bernoulli(0.2),
		Bernoulli(0.8),
	}

	bandit, err := SoftmaxNew(len(arms), τ)
	if err != nil {
		t.Fatalf(err.Error())
	}

	sim, err := MonteCarlo(sims, trials, arms, bandit)
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := sims * trials
	if got := len(sim.Selected); got != expected {
		t.Fatalf("incorrect number of trials: %d", got)
	}
}
