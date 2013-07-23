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

This constructs a bandit with 2 arms using `EpsilonGreedy` with `ε` set to
10%.

## Simulation

The following code runs a monte carlo simulation with the epsilon greedy
algorithm. There are 4 arms with the probability of a reward of 1.0 defined in
`arms`. Results are returned in `Simulation`, which holds a full recording of
the simulation. It can be summarized with the functions `Performance`,
`Accuracy` and `Cumulative`.

```go
import (
  "github.com/purzelrakete/bandit"
  "log"
)

sims := 1000
trials := 400
arms := []bandit.Arm{
  bandit.Bernoulli(0.1),
  bandit.Bernoulli(0.3),
  bandit.Bernoulli(0.2),
  bandit.Bernoulli(0.8),
})

s, err := bandit.MonteCarlo(sims, trials, arms, func() (bandit.Bandit, error) {
  return bandit.EpsilonGreedyNew(len(arms), ε)
})

if err != nil {
  log.Fatalf(err.Error())
}

performance := bandit.Performance(s, 4)
accuracy := bandit.Accuracy(s)
cumulative := bandit.Cumulative(s)
```

## Plotting

You can run and plot a Monte Carlo simulation using the `plot` binary. It will
display the accuracy, performance and cumulative performance over time.

```sh
$GOPATH/bin/plot
open bandit*.png
```

You can change the default number and parameterization of bernoulli arms like
this:

```sh
$GOPATH/bin/plot -mus 0.22,0.1,0.7
open bandit*.png
```

View defaults and available flags:

```sh
$GOPATH/bin/plot -h
```

## Algorithms

- EpsilonGreedy

Randomly explores with a probability of `ε`. The rest of the time, the best
known arm is selected.

- Softmax

Explores arms proportionally to their success. Explore exploit is traded off
by temperature parameter τ. As τ → ∞, the bandit explores randomly. When
τ = 0, the bandit will always explore the best known arm.

