# Bandit

[![Build Status](https://travis-ci.org/purzelrakete/bandit.png?branch=master)](https://travis-ci.org/purzelrakete/bandit)
[![Coverage Status](https://coveralls.io/repos/purzelrakete/bandit/badge.png)](https://coveralls.io/r/purzelrakete/bandit)

A mulitarmed bandit to A/B test go projects, or other languages via an HTTP
API. It uses a log based data flow. Based on John Myles White's [Bandit
Algorithms for Website
Optimization](http://shop.oreilly.com/product/0636920027393.do). Full
documentation is available [on
godoc](http://godoc.org/github.com/purzelrakete/bandit).

Build bandit with `make`. You need >= go 1.1.1..

## Try Bandit

`bandit-example` runs a toy demonstration of the HTTP API which you can see at
http://localhost:8080:

![example](http://goo.gl/oaCF3o)

## Data Flow

Bandit operates a delayed bandit with log data. This means that reward data
(ie clickthroughs) do not reach the bandit in real time. Instead they are
aggregated into snapshots by bandit-job. Each bandit instance then polls for
this snapshot periodically.

```
              select
  bandit     ----->       log       --->  bandit-job
  instance    reward      storage         perodically writes
    ^        ----->                            |
    |                                          |
    .-----------------  snapshot <-------------.
        bandit polls
```

`bandit-job` expects log lines in the following format:

```
1379257984 BanditSelection shape-20130822:1
1379257987 BanditReward shape-20130822:1 0.000000
```

Notice that the reward line includes the variant Tag. It is up to you to
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
inside an Experiment. You set up experiments (ie signup form buttons) with as
many variants as you like (ie blue button, red button):

```
                          .------------.
       .--------.         |  snapshot  |       periodic job
       | Bandit | <------ |    file    | <---  aggregates logs
       .--------.         .------------.       into counters
           ^
           |
      .------------.       .----------.
      | Experiment | --*-> |  Variant |
      .------------.       .----------.
      | name       |       | tag      |
      .------------.       | url      |
                           .----------.
```

## Running experiments

To use a bandit, you first have to define an experiment and it's variants.
This is currently configured as a TSV with name, url, tag. See experiments.tsv
for an example.

Choose the best integration for your project depending on whether you have
a client side javascript application, a go project, or a project in some other
language.

### Javascript and the HTTP API

Run `bandit-api -port 80 -apiExperiments experiments.tsv` to start the
endpoint with the provided test experiments.

In this scenario, the application makes a request to the api endpoint and
then a second request to your api.

```
   .--------------.        .--------------------------.
   |  javascript  | -----> | bandit HTTP API (select) |
   .--------------.        .--------------------------.
          |                .------------.
          ---------------> |  your API  |
          |                .------------.
          |                .---------------------------.
          ---------------> |  bandit HTTP API (reward) |
                           .---------------------------.
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

See the exampe binary and example/index.html for a running example of this.

### Project in another language using the HTTP API

Launch the HTTP API as above. When you get a request to your endpoint, make
a backend request to the HTTP API. Use the returned variant to vary.

### Running experiments in go with the bandit library

Integrate with as follows:

1. Load an experiment.
2. Initialize your own variant code if necessary.
3. Serve. In each request, select a variant with the experiment and serve it.

You can load an experiment with an associated bandit as an Experiment:

```go
e, err := bandit.NewExperiment("experiments.tsv", "shape-20130822")
if err != nil {
  log.Fatalf("could not construct experiment: %s", err.Error())
}

opener := bandit.NewFileOpener("shape-20130822.dsv")
if err := e.InitDelayedBandit(opener, 3 * time.Hours); err != nil {
  log.Fatalf("could initialize bandits: %s", err.Error())
}

fmt.Println(e.Variants)
```

You can iterate over available Variants in your endpoint setup via e.Variants.
Use this to intialize your viariants if you wish. You could also switch
directly on tags:

```go
var msg string
switch e.Select().Tag {
  case "shape-20130822:1":
    msg = "hello square"
  case "shape-20130822:2":
    msg = "hello circle"
}
```

Your response must include the Tag somwhere so the client can tag subsequent
rewards. If you do not do this, you will not be able to calculate rewards.

## Aggregating Logs

In a production setting logs are aggregated as described in "Data Flow". You
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

Thanks to for advice and opinions to

- John Myles White
- Josh Devins
- Ozgür Demir
- Peter Bourgon
- Sean Braithwaite

[1]: http://dl.acm.org/citation.cfm?id=1677012" "Explore/Exploit Schemes for Web Content Optimzation"
