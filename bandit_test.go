// Copyright 2013 SoundCloud, Rany Keddo. All rights reserved.  Use of this
// source code is governed by a license that can be found in the LICENSE file.

package bandit

import (
	//"fmt"
	"github.com/purzelrakete/bandit/sim"
	"math"
	"testing"
)

func TestDrawCategorial(t *testing.T) {
	{
		bins := []float64{0.1, 0.4, 0.05, 0.2, 0.25}
		testX := []float64{0, 0.05, 0.2, 0.5, 0.55, 0.7, 0.8, 1, 5}
		expected := []int{0, 0, 1, 1, 2, 3, 4, 4, -1}

		for idx, value := range expected {
			if got := getBinId(testX[idx], bins); got != value {
				t.Fatalf("incorrect bin selected for x=%f: %d. is %d", testX[idx], got, value)
			}
		}
	}
	// sanity check: simulate drawing from categorial dist
	{
		bins := []float64{0.1, 0.4, 0.05, 0.2, 0.25}
		numReps := 10000
		distCat := categorialDistribution{bins, 1}
		got := make([]int, len(bins))
		for rep := 1; rep < numReps; rep++ {
			got[CatRand(distCat)-1]++ // TODO(cs): use fixed seed
		}

		for idx, value := range bins {
			gotNormed := float64(got[idx]) / float64(numReps)
			if !(gotNormed-value < 0.05) {
				t.Fatalf("empirical ratio converge to %f. is %f", gotNormed, value)
			}
		}
	}
}

func TestSoftmaxCalc(t *testing.T) {
	// compute simple softmax probs
	{
		scores := []float64{math.Log(2), math.Log(3), math.Log(1), math.Log(4)}
		expected := []float64{0.2, 0.3, 0.1, 0.4}
		got := calcSoftMax(scores, 1)

		// check arity of result
		if len(expected) != len(got.individualProbs) {
			t.Fatalf("incorrect length: %d", len(got.individualProbs))
		}

		for i, value := range expected {
			if !(math.Abs(got.individualProbs[i]/got.normalizer-value) < 0.001) {
				t.Fatalf("probability %d should be %f. is %f", i+1, value, got.individualProbs[i]/got.normalizer)
			}
		}
	}

	// check different tau
	{
		scores := []float64{math.Log(2), math.Log(3), math.Log(1), math.Log(4)}
		expected := []float64{0.247146, 0.257373, 0.230595, 0.264884}
		got := calcSoftMax(scores, 10)

		for i, value := range expected {
			if !(math.Abs(got.individualProbs[i]/got.normalizer-value) < 0.001) {
				t.Fatalf("probability %d should be %f. is %f", i+1, value, got.individualProbs[i]/got.normalizer)
			}
		}
	}
	// check huge tau
	{
		scores := []float64{math.Log(2), math.Log(3), math.Log(1), math.Log(4)}
		expected := []float64{0.25, 0.25, 0.25, 0.25}
		got := calcSoftMax(scores, 10000)

		for i, value := range expected {
			if !(math.Abs(got.individualProbs[i]/got.normalizer-value) < 0.001) {
				t.Fatalf("probability %d should be %f. is %f", i+1, value, got.individualProbs[i]/got.normalizer)
			}
		}
	}
	// check tiny tau
	{
		scores := []float64{math.Log(2), math.Log(3), math.Log(1), math.Log(4)}
		expected := []float64{0, 0, 0, 1}
		got := calcSoftMax(scores, 0.0001)

		for i, value := range expected {
			if !(math.Abs(got.individualProbs[i]/got.normalizer-value) < 0.001) {
				t.Fatalf("probability %d should be %f. is %f", i+1, value, got.individualProbs[i]/got.normalizer)
			}
		}
	}

	// check huge scores
	{
		scores := []float64{5000, 1000}
		expected := []float64{1, 0}
		got := calcSoftMax(scores, 1)

		for i, value := range expected {
			if !(math.Abs(got.individualProbs[i]/got.normalizer-value) < 0.001) {
				t.Fatalf("probability %d should be %f. is %f", i+1, value, got.individualProbs[i]/got.normalizer)
			}
		}
	}
}

func TestEpsilonGreedy(t *testing.T) {
	ε := 0.1
	sims := 5000
	trials := 300
	bestArmIndex := 4 // Bernoulli(bestArm)
	bestArm := 0.8
	arms := []sim.Arm{
		sim.Bernoulli(0.1),
		sim.Bernoulli(0.3),
		sim.Bernoulli(0.2),
		sim.Bernoulli(bestArm),
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

	// TODO(cs): change got < 0.9 to !(got > 0.9), such that test fails, if got is nan
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
		sim.Bernoulli(0.1),
		sim.Bernoulli(0.3),
		sim.Bernoulli(0.2),
		sim.Bernoulli(0.8),
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

// TODO(cs): rethink test (suffers from conceptual weakness)
func TestSoftmaxGaussian(t *testing.T) {
	τ := 0.1
	sims := 5000
	trials := 300
	bestArmIndex := 1 // Gaussian(bestArm)
	bestArm := 5000.0
	arms := []sim.Arm{
		sim.Gaussian(5000,1), // is five times better
		sim.Gaussian(1000,1),
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
	fmt.Println(s.Selected)
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
		sim.Bernoulli(0.1),
		sim.Bernoulli(0.3),
		sim.Bernoulli(0.2),
		sim.Bernoulli(0.8),
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
