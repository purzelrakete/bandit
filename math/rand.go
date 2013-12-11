package math

import (
	"math"
	"math/rand"
	"time"
)

// BetaRnd returns beta distributed random variables: x ~ Beta(α, β)
// implementation follows R.C.H. Cheng: Generating Beta Variates with Nonintegral Shape Parameters
func BetaRnd() func(α, β float64) float64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	log4 := math.Log(4)
	return func(α, β float64) float64 {
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
			U1, U2 := r.Float64(), r.Float64()
			V := b * math.Log(U1/(1-U1))
			W = α * math.Exp(V)

			reject = (a*math.Log(a/(β+W)) + c*V - log4) < math.Log(U1*U1*U2)
		}
		return (W / (β + W))
	}
}

// NormRnd returns normally distributed random variables: x ~ N(x|μ,σ)
func NormRnd(μ, σ float64) func() float64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return func() float64 {
		return r.NormFloat64()*σ + μ
	}
}

// DiracRnd returns a constant value c
func DiracRnd(c float64) func() float64 {
	return func() float64 {
		return c
	}
}

// BernRnd returns Bernoulli distributed random variables: x ~ Bern(x|μ)
func BernRnd(μ float64) func() float64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return func() float64 {
		res := 0.0
		if r.Float64() <= μ {
			res = 1.0
		}
		return res
	}
}
