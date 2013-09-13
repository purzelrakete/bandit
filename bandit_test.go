// Copyright 2013 SoundCloud, Rany Keddo. All rights reserved.  Use of this
// source code is governed by a license that can be found in the LICENSE file.

package bandit

import (
	"math"
	"testing"
)

func TestEpsilonGreedy(t *testing.T) {
	ε := 0.1
	sims := 5000
	trials := 300
	bestArmIndex := 4 // Bernoulli(bestArm)
	bestArm := 0.8
	arms := []Arm{
		Bernoulli(0.1),
		Bernoulli(0.3),
		Bernoulli(0.2),
		Bernoulli(bestArm),
	}

	bandit, err := NewEpsilonGreedy(len(arms), ε)
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

func TestSoftmax(t *testing.T) {
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

	bandit, err := NewSoftmax(len(arms), τ)
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

func TestUCB1(t *testing.T) {
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

	sim, err := MonteCarlo(sims, trials, arms, NewUCB1(len(arms)))
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := sims * trials
	if got := len(sim.Selected); got != expected {
		t.Fatalf("incorrect number of trials: %d", got)
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
