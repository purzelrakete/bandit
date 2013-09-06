# Bandit

[![Build Status](https://travis-ci.org/purzelrakete/bandit.png?branch=master)](https://travis-ci.org/purzelrakete/bandit)
[![Coverage Status](https://coveralls.io/repos/purzelrakete/bandit/badge.png)](https://coveralls.io/r/purzelrakete/bandit)

A golang multiarmed bandit. Use it in your project to run A/B tests while
controlling the tradeoff between exploring new arms and exploiting the
currently best arm. It can be used inside a go project or from other languages
through an HTTP API.

Bandit is based on John Myles White's [Bandit
Algorithms for Website Optimization](http://shop.oreilly.com/product/0636920027393.do).
It works well on high volume websites or APIs.

Full documentation is available [on godoc](http://godoc.org/github.com/purzelrakete/bandit).

## Try Bandit

You need at go1.1.1 or higher. Build the project by running `make`.

You can run a simple demonstration of the HTTP API with `$GOPATH/bin/example`.
Go to http://localhost:8080/ to test the performance of squares against
circles. If you perfer circles, you should start to see more circles being
served to you over time.

## Where to use bandit

This library is intended to be used to instrument a high volume website or
a web api. Bandit expects to find snapshots of log aggregations, so you need
to have a single source of logs. If you don't have a logging pipeline you can
use some unix tools to prepare snashots, see below.

## Running experiments

Choose the best method for your project depending on whether you have a client
side javascript application, a go project, or a project in some other
language.

### Javascript and the HTTP API

Run `$GOPATH/bin/oob -port 80 -oobExperiments experiments.tsv` to start the
endpoint with the provided test experiments.

In this scenario, the application makes a request to the api endpoint and
then a second request to your api.

```
  --------------         -----------------
 |  javascript  | ----> | bandit HTTP API |
  --------------         -----------------
        |                 ------------
        ---------------> |  your api  |
                          ------------
```

Get a variant from the HTTP API first:

    GET https://api/test/widgets?uid=11 HTTP/1.0

And receives a json response response

    HTTP/1.0 200 OK
    Content-Type: text/json

    {
      uid: 11,
      experiment: "widgets",
      url: "https://api/widget?color=blue"
      tag: "widget-sauce-flf89"
    }

The client can now follow up with a request to the returned widget:

    GET https://api/widget?color=blue HTTP/1.0

See example/index.html for an example of this.

### Project in another language using the HTTP API

Launch the HTTP API as above. When you get a request to your endpoint, make
a backend request to the HTTP API. Use the returned variant information to
make your response.

```
           ------------         -----------------
    ----> |  your api  | ----> | bandit HTTP API |
           ------------         -----------------
```
### Running experiments as a go library

```
       ------------
----> |  your api  |
       --|bandit|--
```

Set your handler up with a bandit.Test

```go
t, err := bandit.NewDelayedTests(experimentsFile, snapshotFile, 1*time.Minute)
if err != nil {
  log.Fatalf("could not construct experiments: %s", err.Error())
}
```

## Bandit Algorithms

You can currently choose between Epsilon Greedy and Softmax. See the godoc for
detailed information.

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

## TODO

- UCB1 implementation
- UCB with extensions for delayed rewards
- Sticky assignements

# Credits

Developed by

- Rany Keddo (@purzelrakete)

Thanks to for advice and opinions to

- John Myles White
- Peter Burgeon
- Josh Devins

[1]: http://dl.acm.org/citation.cfm?id=1677012" "Explore/Exploit Schemes for Web Content Optimzation"
