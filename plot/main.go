package main

import (
	"flag"
	"github.com/purzelrakete/bandit"
	"log"
)

var (
	mcSims    = flag.Int("mcSims", 5000, "monte carlo simulations to run")
	mcHorizon = flag.Int("mcHorizon", 300, "trials per simulation")
	mcMus     = flag.String("mcMus", "0.1,0.3,0.2,0.8", "bernoulli μs")
)

func init() {
	flag.Parse()
}

func main() {
	μs, bestArms, err := parseArms(*mcMus)
	if err != nil {
		log.Fatal(err.Error())
	}

	// bernoulli arms. these determine the observed distribution
	arms := arms{}
	for _, μ := range μs {
		arms = append(arms, bandit.Bernoulli(μ))
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

	// draw groups
	for _, group := range groups {
		s, err := simulate(group.bandits, arms, *mcSims, *mcHorizon)
		if err != nil {
			log.Fatal(err.Error())
		}

		graph := summarize(s, bandit.Accuracy(bestArms))
		draw(graph, group.name+" Accuracy", "Time", "P(selecting best arm)")

		s, err = simulate(group.bandits, arms, *mcSims, *mcHorizon)
		if err != nil {
			log.Fatal(err.Error())
		}

		graph = summarize(s, bandit.Performance)
		draw(graph, group.name+" Performance", "Time", "P(selecting best arm)")

		s, err = simulate(group.bandits, arms, *mcSims, *mcHorizon)
		if err != nil {
			log.Fatal(err.Error())
		}

		graph = summarize(s, bandit.Cumulative)
		draw(graph, group.name+" Cumulative", "Time", "P(selecting best arm)")
	}
}
