# Bandit

[![Build Status](https://travis-ci.org/purzelrakete/bandit.png?branch=master)](https://travis-ci.org/purzelrakete/bandit)
[![Coverage Status](https://coveralls.io/repos/purzelrakete/bandit/badge.png)](https://coveralls.io/r/purzelrakete/bandit)

A mulitarmed bandit to A/B test go projects, or other languages via an HTTP API.
It uses a log-based data flow.
Based on John Myles White's [Bandit Algorithms for Website Optimization](http://shop.oreilly.com/product/0636920027393.do).
Full documentation is available [on godoc](http://godoc.org/github.com/purzelrakete/bandit).

You can see a general introduction to [Multiarmed Bandits] [1] here.

Build bandit with `make`. You need go >= 1.1.1..

## Data Flow

A bandit instance is embedded into e.g. an HTTP server. Incoming requests select
a variation from an experiment, and then log that selection. Subsequent positive
feedback from that selection, e.g. a click, is also logged.

Periodically, `bandit-job` aggregates selections and rewards from the logs, re-
calculates variation distribution, and emits a snapshot file to some  shared
storage. The bandit polls for updates to that snapshot file and hot-reloads the
distribution on change.

```
+----------+           +----------+                +------------+
| Bandit   |--select-->| Log      |- - periodic - >| bandit-job |
| instance |--reward-->| storage  |                |            |
+----------+           +----------+                +------------+
     ^                                                    |
     |                 +----------+                       |
     '---------poll----| Snapshot |<----------------------'
                       +----------+
```

`bandit-job` expects log lines in the following format:

```
1379257984 BanditSelection shape-20130822:1:8932478932
1379257987 BanditReward shape-20130822:1:8932478932 0.000000
```

Notice that the reward line includes the variation tag. It is up to you to
transport this tag through your system.

## Types

A Strategy is used to select arms and update arms with reward information:

```go
type Strategy interface {
  SelectArm() int
  Update(arm int, reward float64)
}
```

You will probably not use bandits directly. Instead, a Strategy is put to work
inside an Experiment. You set up experiments (e.g. signup form buttons) with as
many variations as you like (e.g. blue button, red button):

```
  +--------+            +---------------+      periodic job
  | Strategy | 1 <----- 1 | Snapshot file | <--- aggregates logs
  +--------+            +---------------+      into counters
      1
      ^
      |
      1
+------------+         +---------+
| Experiment | 1 --> * | Variation |
|------------|         |---------|
| name       |         | tag     |
+------------+         | url     |
                       +---------+
```

## Integrating and running experiments

To use a strategy, you first have to define an experiment and its variations. This
is currently configured in json with a name, URL, and tag. See
experiments.json for an example.

Choose the best integration for your project depending on whether you have
a client side javascript application, a go project, or a project in some other
language.

### Integration with Javascript and the HTTP API

Run `bandit-api -port 80 -apiExperiments experiments.json` to start the
endpoint with the provided test experiments.

In this scenario, the application makes a request to the API endpoint and
then a second request to your API.

```
             Bandit HTTP API
Javascript   Select   Reward   Your API
----------   ------   ------   --------
     |-------->|                  |
     |<--------|                  |
     |                            |
     |--------------------------->|
     |<---------------------------|
     :
   later
     :
     |----------------->|
     |<-----------------|
```

Get a variation from the HTTP API first:

    GET https://api/experiements/widgets?uid=11 HTTP/1.0

The API responds with a variation:

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

See the exampe binary and example/index.html for a running example.

### Integration in another language using the HTTP API

Launch the HTTP API as above. When you get a request to your endpoint, make
a backend request to the HTTP API. Use the returned variation to vary.

### Integration with Go projects

First, load an experiment.

```go
experiments := bandit.NewFileOpener("experiments.json")
e, err := bandit.NewExperiment(experiments, "shape-20130822")
if err != nil {
  log.Fatalf("could not construct experiment: %s", err.Error())
}

fmt.Println(e.Variations)
```

Initialize your own variation code if necessary. Then, serve. In each request,
select a variation via the experiment and serve it. Be sure to include the tag
in the response, so your clients can pass it back with rewards.

# Miscellaneous information

## Aggregating Logs

In a production setting logs are aggregated as described in Data Flow. You
can use `bandit-job` as a streaming map reduce job with `bandit-job -kind map`
and `bandit-job -kind reduce`. You can also run over the logs wiht `bandit-job
-kind poll`. See `bandit-job -h` for information.

## Strategy Algorithms

You can currently choose between Epsilon Greedy, UCB1, Softmax, and Thompson ([see, e.g., Chapelle & Li, 2011 ](http://books.nips.cc/papers/files/nips24/NIPS2011_1232.pdf)). See the
godoc for detailed information.

## Snapshots and delayed bandits

You can configure your strategy to get it's internal state from a snapshot like
this:

[
  {
    "experiment_name": "shape-20130822",
    "strategy": "softmax",
    "parameters": [0.1],
    "snapshot": "snapshot.tsv",
    "snapshot-poll-seconds": 60,
    "variations": [
      {
        "url": "http://localhost:8080/widget?shape=circle",
        "description": "Everybody likes circles.",
        "ordinal": 1
      },
      {
        "url": "http://localhost:8080/widget?shape=square",
        "description": "Everybody likes squares.",
        "ordinal": 2
      }
    ]
  }
]

## Simulation

The `bandit/sim` package includes the facility to simulate and plot
experiemnts. You should run your own simulations before putting experiments
into production. See the sim package for details. You can run bandit-plot
to see some out of the box simulations.

# Status

Version: 0.0.0-alpha.1

The API is currently *not stable* and is subject to change at any time.

## TODO

- UCB with extensions for delayed rewards
- Sticky assignments
- Extend Thompson sampling for different reward distributions

# Credits

Developed by

- Rany Keddo (@purzelrakete)
- Ozgür Demir (@ozgurdemir)
- Christoph Sawade

Thanks to for advice and opinions to

- John Myles White
- Josh Devins
- Peter Bourgon
- Sean Braithwaite

[1]: http://goo.gl/wQkSga "Multiarmed Bandits"
[2]: http://dl.acm.org/citation.cfm?id=1677012 "Explore/Exploit Schemes for Web Content Optimzation"
