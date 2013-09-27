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
a variant from an experiment, and then log that selection. Subsequent positive
feedback from that selection, e.g. a click, is also logged.

Periodically, `bandit-job` aggregates selections and rewards from the logs, re-
calculates variant distribution, and emits a snapshot file to some  shared
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
1379257984 BanditSelection shape-20130822:1
1379257987 BanditReward shape-20130822:1 0.000000
```

Notice that the reward line includes the variant tag. It is up to you to
transport this tag through your system.

## Types

A Bandit is used to select arms and update arms with reward information:

```go
type Bandit interface {
  SelectArm() int
  Update(arm int, reward float64)
}
```

You will probably not use bandits directly. Instead, a Bandit is put to work
inside an Experiment. You set up experiments (e.g. signup form buttons) with as
many variants as you like (e.g. blue button, red button):

```
  +--------+            +---------------+      periodic job
  | Bandit | 1 <----- 1 | Snapshot file | <--- aggregates logs
  +--------+            +---------------+      into counters
      1
      ^
      |
      1
+------------+         +---------+
| Experiment | 1 --> * | Variant |
|------------|         |---------|
| name       |         | tag     |
+------------+         | url     |
                       +---------+
```

## Integrating and running experiments

To use a bandit, you first have to define an experiment and its variants. This
is currently configured as a TSV with a name, URL, and tag. See experiments.tsv
for an example.

Choose the best integration for your project depending on whether you have
a client side javascript application, a go project, or a project in some other
language.

### Integration with Javascript and the HTTP API

Run `bandit-api -port 80 -apiExperiments experiments.tsv` to start the
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

Get a variant from the HTTP API first:

    GET https://api/experiements/widgets?uid=11 HTTP/1.0

The API responds with a variant:

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
a backend request to the HTTP API. Use the returned variant to vary.

### Integration with Go projects

First, load an experiment.

```go

experiments := bandit.NewFileOpener("experiments.tsv")
e, err := bandit.NewExperiment(experiments, "shape-20130822")
if err != nil {
  log.Fatalf("could not construct experiment: %s", err.Error())
}

snapshot := bandit.NewFileOpener("shape-20130822.tsv")
if err := e.InitDelayedBandit(snapshot, 3 * time.Hours); err != nil {
  log.Fatalf("could initialize bandits: %s", err.Error())
}

fmt.Println(e.Variants)
```

Initialize your own variant code if necessary. Then, serve. In each request,
select a variant via the experiment and serve it. Be sure to include the tag
in the response, so your clients can pass it back with rewards.

# Miscellaneous information

## Aggregating Logs

In a production setting logs are aggregated as described in Data Flow. You
can use `bandit-job` as a streaming map reduce job with `bandit-job -kind map`
and `bandit-job -kind reduce`. You can also run over the logs wiht `bandit-job
-kind poll`. See `bandit-job -h` for information.

## Bandit Algorithms

You can currently choose between Epsilon Greedy, UCB1 and Softmax. See the
godoc for detailed information.

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
- Sticky assignements

# Credits

Developed by

- Rany Keddo (@purzelrakete)
- Ozgür Demir (@ozgurdemir)

Thanks to for advice and opinions to

- John Myles White
- Josh Devins
- Peter Bourgon
- Sean Braithwaite

[1]: http://goo.gl/wQkSga "Multiarmed Bandits"
[2]: http://dl.acm.org/citation.cfm?id=1677012 "Explore/Exploit Schemes for Web Content Optimzation"
