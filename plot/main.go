// Copyright 2013 SoundCloud, Rany Keddo. All rights reserved.  Use of this
// source code is governed by a license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"github.com/purzelrakete/bandit"
	"github.com/purzelrakete/bandit/math"
	"github.com/purzelrakete/bandit/sim"
	"log"
	"strconv"
	"strings"
)

var (
	mcSims    = flag.Int("mcSims", 5000, "monte carlo simulations to run")
	mcHorizon = flag.Int("mcHorizon", 300, "trials per simulation")
	mcMus     = flag.String("mcMus", "0.1,0.3,0.2,0.8", "bernoulli μs")
)

func init() {
	flag.Parse()
}

// You can run and plot a Monte Carlo simulation using the `plot` binary. It
// will display the accuracy, performance and cumulative performance over
// time.
//
// You can change the default number and parameterization of bernoulli arms
// like this:
//
// plot -mus 0.22,0.1,0.7
//
func main() {
	μs, bestArms, err := parseArms(*mcMus)
	if err != nil {
		log.Fatal(err.Error())
	}

	// bernoulli arms. this is the hidden distribution.
	arms := arms{}
	for _, μ := range μs {
		arms = append(arms, math.BernRand(μ))
	}

	// groups of graphs to draw
	groups := []group{}

	// epsilon greedy
	greedys := bandits{}
	for _, ε := range []float64{0.1, 0.2, 0.3, 0.4, 0.5} {
		bandit, err := bandit.NewEpsilonGreedy(len(μs), ε)
		if err != nil {
			log.Fatal(err.Error())
		}

		greedys = append(greedys, bandit)
	}

	groups = append(groups, group{
		name:    "Epsilon Greedy",
		bandits: greedys,
	})

	// softmax
	softmaxes := bandits{}
	for _, τ := range []float64{0.1, 0.2, 0.3, 0.4, 0.5} {
		bandit, err := bandit.NewSoftmax(len(μs), τ)
		if err != nil {
			log.Fatal(err.Error())
		}

		softmaxes = append(softmaxes, bandit)
	}

	groups = append(groups, group{
		name:    "Softmax",
		bandits: softmaxes,
	})

	// ucb1
	ucb1s := bandits{
		bandit.NewUCB1(len(μs)),
	}

	groups = append(groups, group{
		name:    "UCB1",
		bandits: ucb1s,
	})

	// thompson sampling
	thompsons := bandits{}
	for _, α := range []float64{1, 2, 10, 20, 100} {
		bandit, err := bandit.NewThompson(len(μs), α)
		if err != nil {
			log.Fatal(err.Error())
		}

		thompsons = append(thompsons, bandit)
	}

	groups = append(groups, group{
		name:    "Thompson Sampling",
		bandits: thompsons,
	})

	// mixed
	mixed := bandits{}
	greedy, err := bandit.NewEpsilonGreedy(len(μs), 0.1)
	if err != nil {
		log.Fatal(err.Error())
	}

	mixed = append(mixed, greedy)

	softmax, err := bandit.NewSoftmax(len(μs), 0.1)
	if err != nil {
		log.Fatal(err.Error())
	}

	mixed = append(mixed, softmax)

	// ucb1 into mixed
	mixed = append(mixed, bandit.NewUCB1(len(μs)))

	groups = append(groups, group{
		name:    "Comparative",
		bandits: mixed,
	})

	// thompson into mixed
	thompson, err := bandit.NewThompson(len(μs), 10.0)
	if err != nil {
		log.Fatal(err.Error())
	}

	mixed = append(mixed, thompson)

	groups = append(groups, group{
		name:    "Comparative",
		bandits: mixed,
	})

	// draw groups
	for _, group := range groups {
		s, err := simulate(group.bandits, arms, *mcSims, *mcHorizon)
		if err != nil {
			log.Fatal(err.Error())
		}

		graph := summarize(s, sim.Accuracy(bestArms))
		draw(graph, group.name+" Accuracy", "Time", "P(selecting best arm)")

		s, err = simulate(group.bandits, arms, *mcSims, *mcHorizon)
		if err != nil {
			log.Fatal(err.Error())
		}

		graph = summarize(s, sim.Performance)
		draw(graph, group.name+" Performance", "Time", "Average Reward")

		s, err = simulate(group.bandits, arms, *mcSims, *mcHorizon)
		if err != nil {
			log.Fatal(err.Error())
		}

		graph = summarize(s, sim.Cumulative)
		draw(graph, group.name+" Cumulative", "Time", "Cumulative Reward")
	}
}

// parseArms converts command line 0.1,0.2 into a slice of floats. Returns
// the the best arm (1 indexed). In the case of equally good best arms there
// will be multiple indices in the returned slice.
func parseArms(sμ string) ([]float64, []int, error) {
	var μs []float64
	var imax []int
	max := 0.0
	for i, s := range strings.Split(sμ, ",") {
		μ, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return []float64{}, []int{}, fmt.Errorf("not a number: %s", err.Error())
		}

		if μ < 0 || μ > 1 {
			return []float64{}, []int{}, fmt.Errorf("μ not in [0,1]: %.5f", μ)
		}

		// there may be multiple equally good (best) arms
		if μ > max {
			max = μ
			imax = []int{i + 1}
		} else if μ == max {
			imax = append(imax, i+1)
		}

		μs = append(μs, μ)
	}

	return μs, imax, nil
}
