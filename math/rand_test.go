package math

import (
	"math"
	"testing"
)

func TestBetaRand(t *testing.T) {
	var seed int64 = 123 //time.Now().UnixNano()
	betaRnd := NewBetaRand(seed)
	α, β := 15.0, 4.0
	numSamples := 1000000
	expectation := α / (α + β)
	variance := α * β / ((α + β) * (α + β) * (α + β + 1))

	mean, mean2 := 0.0, 0.0
	for i := 0; i < numSamples; i++ {
		x := betaRnd.NextBeta(α, β)
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

func TestBetaSeed(t *testing.T) {
	var seed int64 = 123 //time.Now().UnixNano()
	betaRnd := NewBetaRand(seed)
	α, β := 15.0, 4.0
	expected := 0.810981

	if got := betaRnd.NextBeta(α, β); math.Abs(got-expected) > 0.0000001 {
		t.Fatalf("beta random variable should be %f, but is %f", expected, got)
	}
}
