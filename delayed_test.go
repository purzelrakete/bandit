package bandit

import (
	"math"
	"testing"
)

func TestDelayedBandit(t *testing.T) {
	τ := 0.1
	sims := 5000
	trials := 300
	flushAfter := 100
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

	d := NewSimulatedDelayedBandit(b, len(arms), flushAfter)

	sim, err := MonteCarlo(sims, trials, arms, d)
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
