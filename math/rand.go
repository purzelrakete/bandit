package math

import (
	"math"
	"math/rand"
	"time"
)

// A BetaRand is a source of random numbers.
type BetaRand struct {
	rand *rand.Rand // seeded random number generator to generate other random values.
}

// NewBetaRand returns a new BetaRand that uses random values from rand
// to generate beta random values.
func NewBetaRand(seed int64) *BetaRand {
	return &BetaRand{rand.New(rand.NewSource(seed))}
}

// NextBeta returns beta distributed random variables: x ~ Beta(α, β)
// implementation follows R.C.H. Cheng: Generating Beta Variates with Nonintegral Shape Parameters
func (r *BetaRand) NextBeta(α, β float64) float64 {
	// initialization
	a := α + β
	b := math.NaN()
	if math.Min(α, β) <= 1 {
		b = math.Max(1/α, 1/β)
	} else {
		b = math.Sqrt((a - 2) / (2*α*β - a))
	}
	c := α + 1/b

	// start rejection sampling
	W := math.NaN()
	for reject := true; reject; {
		U1, U2 := r.rand.Float64(), r.rand.Float64()
		V := b * math.Log(U1/(1-U1))
		W = α * math.Exp(V)

		reject = (a*math.Log(a/(β+W)) + c*V - math.Log(4)) < math.Log(U1*U1*U2)
	}
	return (W / (β + W))
}

// NormRand returns normally distributed random variables: x ~ N(x|μ,σ)
func NormRand(μ, σ float64) func() float64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return func() float64 {
		return r.NormFloat64()*σ + μ
	}
}

// DiracRand returns a constant value c
func DiracRand(c float64) func() float64 {
	return func() float64 {
		return c
	}
}

// BernRand returns Bernoulli distributed random variables: x ~ Bern(x|μ)
func BernRand(μ float64) func() float64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return func() float64 {
		res := 0.0
		if r.Float64() <= μ {
			res = 1.0
		}
		return res
	}
}
