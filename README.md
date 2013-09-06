# Bandit

[![Build Status](https://travis-ci.org/purzelrakete/bandit.png?branch=master)](https://travis-ci.org/purzelrakete/bandit)
[![Coverage Status](https://coveralls.io/repos/purzelrakete/bandit/badge.png)](https://coveralls.io/r/purzelrakete/bandit)

A golang multiarmed bandit. Use it in your project to run A/B tests while
controlling the tradeoff between exploring new arms and exploiting the
currently best arm. It can be used inside a go project or from other languages
through an HTTP API and works well on high volume websites or APIs.

Bandit is based on John Myles White's [Bandit
Algorithms for Website Optimization](http://shop.oreilly.com/product/0636920027393.do).

Full documentation is available [on godoc](http://godoc.org/github.com/purzelrakete/bandit).

## Try Bandit

You need at go1.1.1 or higher. Build the project by running `make`.

You can run a simple demonstration of the HTTP API with `$GOPATH/bin/example`.
Go to http://localhost:8080/ to test the performance of squares against
circles. If you perfer circles, you should start to see more circles being
served to you over time.

## When to use bandit

This library is intended to be used to instrument a high volume website or
a web api. It is helpful to have a logging pipeline in place.

## Design

A Bandit is used to select arms and update arms with reward information:

```go
type Bandit interface {
  SelectArm() int
  Update(arm int, reward float64)
}
```

A delayed Bandit has no Update implementation. Instead it maintains static
rewards counters for each arm, and periodically updates those counters from an
externally generated snapshot.  This snapshot contains the number of arms
followed by the mean reward for each arm:

```
2 0.4 0.3
```

You are expected to generate this reward snapshot yourself using log
information. If you adhere to the provided log format, you can use
bandit.SnapshotMapper and bandit.SnapshotReducer to either run a hadoop
streaming job or simply pipe the two commands together in your shell. Have
a look at `log.go` to see the format:

```
2013/08/22 14:20:05 BanditSelection shape-20130822 0 shape-20130822:c8-circle
2013/08/22 14:20:06 BanditReward shape-20130822 0 shape-20130822:c8-circle 1.0
```

To use a bandit, you first have to define an experiment and it's variants.
This is currently configured as a tsv with name, url, tag:

```
shape-20130822	1	http://localhost:8080/widget?shape=square	shape-20130822:s1-square
shape-20130822	2	http://localhost:8080/widget?shape=circle	shape-20130822:c8-circle
```

## Running experiments

Choose the best method for your project depending on whether you have a client
side javascript application, a go project, or a project in some other
language.

### Javascript and the HTTP API

Run `$GOPATH/bin/api -port 80 -apiExperiments experiments.tsv` to start the
endpoint with the provided test experiments.

In this scenario, the application makes a request to the api endpoint and
then a second request to your api.

```
   .--------------.        .-----------------.
   |  javascript  | -----> | bandit HTTP API |
   .--------------.        .-----------------.
          |                .------------.
          ---------------> |  your api  |
                           .------------.
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
a backend request to the HTTP API. Use the returned variant to vary.

```
          .------------.       .-----------------.
    ----> |  your api  | ----> | bandit HTTP API |
          .------------.       .-----------------.
```

### Running experiments with the go library

You can load an experiment with an associated bandit as a Trial.

```
                                     .------------.
                 .--------.          |  snapshot  |       periodic job
       --------> | Bandit | <------  |    file    | <---  aggregates logs
      |          .--------.          .------------.       into counters
   .-------.
   | Trial |
   .-------.     .------------.       .----------.
      |          | Experiment | --*-> |  Variant |
       --------> .------------.       .----------.
                 | name       |       | tag      |
                 .------------.       | url      |
                                      .----------.
```

Set your handler up with a bandit.Trial

```go
trials, err := bandit.NewDelayedTrials(experiments, snapshot, 1*time.Minute)
if err != nil {
  log.Fatalf("could not set up trial: %s", err.Error())
}

t, ok := trails["shape-20130822"]
if !ok {
  log.Fatalf("could not find campaign")
}

m := pat.New()
m.Get("/widget, MyEndpoint(t))
http.Handle("/", m)

log.Fatal(http.ListenAndServe(*apiBind, nil))
```

You can iterate over available Variants in your endpoint setup via
t.Experiment.Variants. Then begin serving requests. For example:

```
switch t.Select().Tag {
  case "shape-20130822:s1-square":
    msg = "hello square"
  case "shape-20130822:c8-circle":
    msg = "hello circle"
}

```

## Bandit Algorithms

You can currently choose between Epsilon Greedy and Softmax. See the godoc for
detailed information.

## Simulation

Bandit includes the facility to simulate and plot experiemnts. You should run
your own simulations before putting experiments into production. See `mc.go`
for details. Too plot the provided simulations, run $GOPATH/bin/plot. You'll
get something like this:

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
- Sean Braithwaite

[1]: http://dl.acm.org/citation.cfm?id=1677012" "Explore/Exploit Schemes for Web Content Optimzation"
