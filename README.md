# Bandit

[![Build Status](https://travis-ci.org/purzelrakete/bandit.png)](https://travis-ci.org/purzelrakete/bandit)

A golang multiarmed bandit. Runs A/B tests while controlling the tradeoff
between exploring new arms and exploiting the currently best arm.

The project is based on John Myles White's [Bandit
Algorithms for Website Optimization](http://shop.oreilly.com/product/0636920027393.do).

## Api

Bandits are fully defined by the following interface:

```go
type Bandit interface {
  SelectArm() int
  Update(arm int, reward float64)
}
```

You should construct a concrete bandit like this:

```go
b := bandit.EpsilonGreedyNew(2, 0.1)
```

This constructs a bandit with 2 arms using EpsilonGreed with ε set to 10%.

## Simulation

The following code runs a monte carlo simulation with the epsilon greedy
algorithm. There are 4 arms with the probability of a reward of 1.0 defined in
μs. Results are returned in `Simulation`, which holds a full recording of the
simulation. It can be summarized with the functions Performance, Accuracy and
Cumulative.

```go
μs := []float64{0.1, 0.3, 0.2, 0.8}
sims := 1000
trials := 400
banditNew := func() (bandit.Bandit, error) {
  return bandit.EpsilonGreedyNew(len(μs), ε)
}

s, err := bandit.MonteCarlo(sims, trials, banditNew, []bandit.Arm{
  bandit.Bernoulli(μs[0]),
  bandit.Bernoulli(μs[1]),
  bandit.Bernoulli(μs[2]),
  bandit.Bernoulli(μs[3]),
})

if err != nil {
  log.Fatalf(err.Error())
}

performance := bandit.Performance(s, 4)
accuracy := bandit.Accuracy(s)
cumulative := bandit.Cumulative(s)
```

# Plotting

You can run and plot a Monte Carlo simulation using the `plot` binary. It will
display the accuracy, performance and cumulative performance over time.

```sh
$GOPATH/bin/plot
open bandit*.png
```

