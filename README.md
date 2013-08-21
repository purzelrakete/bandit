# Bandit

[![Build Status](https://travis-ci.org/purzelrakete/bandit.png)](https://travis-ci.org/purzelrakete/bandit)
[![Coverage Status](https://coveralls.io/repos/purzelrakete/bandit/badge.png)](https://coveralls.io/r/purzelrakete/bandit)

A golang multiarmed bandit. Runs A/B tests while controlling the tradeoff
between exploring new arms and exploiting the currently best arm.

The project is based on John Myles White's [Bandit
Algorithms for Website Optimization](http://shop.oreilly.com/product/0636920027393.do).

Full documentation is available [on godoc](http://godoc.org/github.com/purzelrakete/bandit).

## Building

Build the project by running `make`. You need at go1.1.1 or higher to build
the project.

## Try

You can run the bundled example application from the project root. It's a
very simple demonstration of the out of band api. Run `$GOPATH/bin example`,
then go to http://localhost:8080/ to test the performance of squares
against circles. If you perfer circles, you should start to see more circles
being served to you over time.

See example/index.html for usage.

## Usage scenario

This library is intended to be used to instrument a website or a web api. This
setting poses some special problems:

* Delayed feedback (batch reward aggregation)
* Non stationary rewards (diurnal cycles)

See [Explore/Exploit Schemes for Web Content Optimzation] [1] for a discussion
of these problems.

## Out of band testing with the HTTP endpoint

The SelectionHandler can be used as an out of the box API endpoint for
javascript applications.

In this scenario, the application makes a request to the api endpoint:

    GET https://api/test/widgets?uid=11 HTTP/1.0

And receives a json response response

    HTTP/1.0 200 OK
    Content-Type: text/json

    {
      uid: 11,
      campaign: "widgets",
      url: "https://api/widget?color=blue"
      tag: "widget-sauce-flf89"
    }

The client can now follow up with a request to the returned widget:

    GET https://api/widget?color=blue HTTP/1.0

### Starting the out of band endpoint

Run `$GOPATH/bin oob -port 80 -campaignFile campaigns.tsv` to start the
endpoint with the provided test campaigns.

## Direct use

Bandits are fully defined by the following interface:

```go
type Bandit interface {
  SelectArm() int
  Update(arm int, reward float64)
  Reset()
  Version() string
}
```

You should construct a concrete bandit like this:

```go
b := bandit.NewSoftmax(2, 0.1)
```

This constructs a bandit with 2 arms using `Softmax` with `τ` set to 0.1.

## Algorithms

### EpsilonGreedy

Randomly explores with a probability of `ε`. The rest of the time, the best
known arm is selected.

### Softmax

Explores arms proportionally to their success. Explore exploit is traded off
by temperature parameter τ. As τ → ∞, the bandit explores randomly. When
τ = 0, the bandit will always explore the best known arm.

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

bandit, err := NewEpsilonGreedy(len(arms), ε)
if err != nil {
  log.Fatal(err.Error())
}

sim, err := MonteCarlo(sims, trials, arms, bandit)
if err != nil {
  log.Fatal(err.Error())
}

performance := bandit.Performance(sim, 4)
ccuracy := bandit.Accuracy(sim)
cumulative := bandit.Cumulative(sim)
```

## Plotting

You can run and plot a Monte Carlo simulation using the `plot` binary. It will
display the accuracy, performance and cumulative performance over time.

```sh
$GOPATH/bin/plot
open bandit*.svg
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

You'll get something like this.

![plot](https://dl.dropboxusercontent.com/u/1704851/bandit.svg)

# Status

This API is currently *not stable*. Consider this a 0.0.0 release that is
subject to change at any time.

# Credits

Developed by

- Rany Keddo (@purzelrakete)

Thanks to John Myles White for some advice.

[1]: http://dl.acm.org/citation.cfm?id=1677012" "Explore/Exploit Schemes for Web Content Optimzation"
