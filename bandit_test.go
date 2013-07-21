package bandit

import (
	"testing"
)

func TestEpsilonGreedy(t *testing.T) {
	ε := 0.1
	sims := 2
	horizon := 5
	μs := []float64{0.1, 0.3, 0.2, 0.8}
	bandit := func() (Bandit, error) { return EpsilonGreedyNew(len(μs), ε) }

	d, err := MonteCarlo(sims, horizon, bandit, []Arm{
		Bernoulli(μs[0]),
		Bernoulli(μs[1]),
		Bernoulli(μs[2]),
		Bernoulli(μs[3]),
	})

	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := sims * horizon
	if got := len(d.Selected); got != expected {
		t.Fatalf("incorrect number of trials: %d", got)
	}
}
