// Copyright 2013 SoundCloud, Rany Keddo. All rights reserved.  Use of this
// source code is governed by a license that can be found in the LICENSE file.

package bandit

import (
	bmath "github.com/purzelrakete/bandit/math"
	"github.com/purzelrakete/bandit/sim"
	"math"
	"testing"
)

func TestEpsilonGreedy(t *testing.T) {
	ε := 0.1
	sims := 5000
	trials := 300
	bestArmIndex := 4 // Bernoulli(bestArm)
	bestArm := 0.8
	arms := []sim.Arm{
		bmath.BernRnd(0.1),
		bmath.BernRnd(0.3),
		bmath.BernRnd(0.2),
		bmath.BernRnd(bestArm),
	}

	bandit, err := NewEpsilonGreedy(len(arms), ε)
	if err != nil {
		t.Fatalf(err.Error())
	}

	s, err := sim.MonteCarlo(sims, trials, arms, bandit)
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := sims * trials
	if got := len(s.Selected); got != expected {
		t.Fatalf("incorrect number of trials: %d", got)
	}

	accuracies := sim.Accuracy([]int{bestArmIndex})(&s)
	if got := accuracies[len(accuracies)-1]; got < 0.9 {
		t.Fatalf("accuracy is only %f. %d sims, %d trials", got, sims, trials)
	}

	performances := sim.Performance(&s)
	if got := performances[len(performances)-1]; math.Abs(bestArm-got) > 0.1 {
		t.Fatalf("performance converge to %f. is %f", bestArm, got)
	}

	expectedCumulative := 200.0
	cumulatives := sim.Cumulative(&s)
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
	arms := []sim.Arm{
		bmath.BernRnd(0.1),
		bmath.BernRnd(0.3),
		bmath.BernRnd(0.2),
		bmath.BernRnd(0.8),
	}

	bandit, err := NewSoftmax(len(arms), τ)
	if err != nil {
		t.Fatalf(err.Error())
	}

	s, err := sim.MonteCarlo(sims, trials, arms, bandit)
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := sims * trials
	if got := len(s.Selected); got != expected {
		t.Fatalf("incorrect number of trials: %d", got)
	}

	accuracies := sim.Accuracy([]int{bestArmIndex})(&s)
	if got := accuracies[len(accuracies)-1]; got < 0.9 {
		t.Fatalf("accuracy is only %f. %d sims, %d trials", got, sims, trials)
	}

	performances := sim.Performance(&s)
	if got := performances[len(performances)-1]; math.Abs(bestArm-got) > 0.1 {
		t.Fatalf("performance converge to %f. is %f", bestArm, got)
	}

	expectedCumulative := 200.0
	cumulatives := sim.Cumulative(&s)
	if got := cumulatives[len(cumulatives)-1]; got < expectedCumulative {
		t.Fatalf("cumulative performance should be > %f. is %f", expectedCumulative, got)
	}
}

func TestSoftmaxGaussian(t *testing.T) {
	τ := 0.1
	sims := 5000
	trials := 300
	bestArmIndex := 1 // Gaussian(bestArm)
	bestArm := 5000.0
	arms := []sim.Arm{
		bmath.NormRnd(5000, 1), // is five times better
		bmath.NormRnd(0, 1),
	}

	bandit, err := NewSoftmax(len(arms), τ)
	if err != nil {
		t.Fatalf(err.Error())
	}

	s, err := sim.MonteCarlo(sims, trials, arms, bandit)
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := sims * trials
	if got := len(s.Selected); got != expected {
		t.Fatalf("incorrect number of trials: %d", got)
	}

	accuracies := sim.Accuracy([]int{bestArmIndex})(&s)
	if got := accuracies[len(accuracies)-1]; got != 1.0 {
		t.Fatalf("accuracy is only %f. %d sims, %d trials", got, sims, trials)
	}

	performances := sim.Performance(&s)
	if got := performances[len(performances)-1]; math.Abs(bestArm-got) > 0.1 {
		t.Fatalf("performance converge to %f. is %f", bestArm, got)
	}

	expectedCumulative := 4500.0 * float64(trials) // (mean(bestArm)-tolerance) * num trials
	cumulatives := sim.Cumulative(&s)
	if got := cumulatives[len(cumulatives)-1]; got < expectedCumulative {
		t.Fatalf("cumulative performance should be > %f. is %f", expectedCumulative, got)
	}
}

func TestUCB1(t *testing.T) {
	sims := 5000
	trials := 300
	bestArmIndex := 4 // Bernoulli(bestArm)
	bestArm := 0.8
	arms := []sim.Arm{
		bmath.BernRnd(0.1),
		bmath.BernRnd(0.3),
		bmath.BernRnd(0.2),
		bmath.BernRnd(0.8),
	}

	s, err := sim.MonteCarlo(sims, trials, arms, NewUCB1(len(arms)))
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := sims * trials
	if got := len(s.Selected); got != expected {
		t.Fatalf("incorrect number of trials: %d", got)
	}

	accuracies := sim.Accuracy([]int{bestArmIndex})(&s)
	if got := accuracies[len(accuracies)-1]; got < 0.9 {
		t.Fatalf("accuracy is only %f. %d sims, %d trials", got, sims, trials)
	}

	performances := sim.Performance(&s)
	if got := performances[len(performances)-1]; math.Abs(bestArm-got) > 0.1 {
		t.Fatalf("performance converge to %f. is %f", bestArm, got)
	}

	expectedCumulative := 200.0
	cumulatives := sim.Cumulative(&s)
	if got := cumulatives[len(cumulatives)-1]; got < expectedCumulative {
		t.Fatalf("cumulative performance should be > %f. is %f", expectedCumulative, got)
	}
}

func TestDelayedBandit(t *testing.T) {
	τ := 0.1
	sims := 5000
	trials := 300
	flushAfter := 100
	bestArmIndex := 4 // Bernoulli(bestArm)
	bestArm := 0.8
	arms := []sim.Arm{
		bmath.BernRnd(0.1),
		bmath.BernRnd(0.3),
		bmath.BernRnd(0.2),
		bmath.BernRnd(0.8),
	}

	b, err := NewSoftmax(len(arms), τ)
	if err != nil {
		t.Fatalf(err.Error())
	}

	d := NewSimulatedDelayedBandit(b, len(arms), flushAfter)

	s, err := sim.MonteCarlo(sims, trials, arms, d)
	if err != nil {
		t.Fatalf(err.Error())
	}

	accuracies := sim.Accuracy([]int{bestArmIndex})(&s)
	if got := accuracies[len(accuracies)-1]; got < 0.9 {
		t.Fatalf("accuracy is only %f. %d sims, %d trials", got, sims, trials)
	}

	performances := sim.Performance(&s)
	if got := performances[len(performances)-1]; math.Abs(bestArm-got) > 0.1 {
		t.Fatalf("performance converge to %f. is %f", bestArm, got)
	}

	expectedCumulative := 200.0
	cumulatives := sim.Cumulative(&s)
	if got := cumulatives[len(cumulatives)-1]; got < expectedCumulative {
		t.Fatalf("cumulative performance should be > %f. is %f", expectedCumulative, got)
	}
}


func TestThompson(t *testing.T) {
	α := 10.0
	sims := 5000
	trials := 300
	bestArmIndex := 4 // Bernoulli(bestArm)
	bestArm := 0.8
	arms := []sim.Arm{
		bmath.BernRnd(0.1),
		bmath.BernRnd(0.3),
		bmath.BernRnd(0.2),
		bmath.BernRnd(0.8),
	}

	bandit, err := NewThompson(len(arms), α)
	if err != nil {
		t.Fatalf(err.Error())
	}

	s, err := sim.MonteCarlo(sims, trials, arms, bandit)
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := sims * trials
	if got := len(s.Selected); got != expected {
		t.Fatalf("incorrect number of trials: %d", got)
	}

	accuracies := sim.Accuracy([]int{bestArmIndex})(&s)
	if got := accuracies[len(accuracies)-1]; got < 0.9 {
		t.Fatalf("accuracy is only %f. %d sims, %d trials", got, sims, trials)
	}

	performances := sim.Performance(&s)
	if got := performances[len(performances)-1]; math.Abs(bestArm-got) > 0.1 {
		t.Fatalf("performance converge to %f. is %f", bestArm, got)
	}

	expectedCumulative := 200.0
	cumulatives := sim.Cumulative(&s)
	if got := cumulatives[len(cumulatives)-1]; got < expectedCumulative {
		t.Fatalf("cumulative performance should be > %f. is %f", expectedCumulative, got)
	}
}
