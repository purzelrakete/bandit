package math

import (
	"math"
	"testing"
)

func TestBetaRnd(t *testing.T) {
	betaRnd := BetaRnd()
	α, β := 15.0, 4.0
	numSamples := 1000000
	expectation := α / (α + β)
	variance := α * β / ((α + β) * (α + β) * (α + β + 1))

	mean, mean2 := 0.0, 0.0
	for i := 0; i < numSamples; i++ {
		x := betaRnd(α, β)
		mean += x
		mean2 += x * x
	}
	mean /= float64(numSamples)
	mean2 /= float64(numSamples)

	// compare mean with expected value
	if math.Abs(mean-expectation) > 0.001 {
		t.Fatalf("mean converge to %f. is %f", expectation, mean)
	}

	// compare sample variance with variance
	if got := mean2 - mean*mean; math.Abs(got-variance) > 0.001 {
		t.Fatalf("variance converge to %f. is %f", variance, got)
	}
}
