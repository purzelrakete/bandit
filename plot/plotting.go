package main

import (
	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"flag"
	"fmt"
	"github.com/purzelrakete/bandit"
	"image/color"
	"log"
)

var (
	mcSims    = flag.Int("mcSims", 5000, "monte carlo simulations to run")
	mcHorizon = flag.Int("mcHorizon", 300, "trials per simulation")
	mcPerfPng = flag.String("mcPerfPng", "performance.png", "performance plot")
)

func init() {
	flag.Parse()
}

func main() {
	εs := []float64{0.1, 0.2, 0.3, 0.4, 0.5}
	μs := []float64{0.1, 0.3, 0.2, 0.8}
	sims := make(map[float64]bandit.Sim)

	for _, ε := range εs {
		banditNew := func() (bandit.Bandit, error) {
			return bandit.EpsilonGreedyNew(len(μs), ε)
		}

		s, err := bandit.MonteCarlo(*mcSims, *mcHorizon, banditNew, []bandit.Arm{
			bandit.Bernoulli(μs[0]),
			bandit.Bernoulli(μs[1]),
			bandit.Bernoulli(μs[2]),
			bandit.Bernoulli(μs[3]),
		})

		if err != nil {
			log.Fatalf(err.Error())
		}

		sims[ε] = s
	}

	p, err := plot.New()
	if err != nil {
		log.Fatalf(err.Error())
	}

	p.Title.Text = "Epsilon Greedy Performance"
	p.X.Label.Text = "Time"
	p.Y.Label.Text = "Average reward"

	for ε, sim := range sims {
		l, err := plotter.NewLine(performance(sim))
		if err != nil {
			log.Fatalf(err.Error())
		}

		p.Add(l)
		p.Legend.Add(fmt.Sprintf("%.2f", ε), l)
		l.LineStyle.Color = color.Gray{uint8(255 * 1.9 * ε)}
	}

	if err != nil {
		log.Fatalf(err.Error())
	}

	if err := p.Save(8, 8, *mcPerfPng); err != nil {
		log.Fatalf(err.Error())
	}
}

// performance averaged over sims at each time point
func performance(s bandit.Sim) plotter.XYs {
	data := bandit.Performance(s)
	points := make(plotter.XYs, len(data))
	for i, datum := range data {
		points[i].X = float64(i)
		points[i].Y = datum
	}

	return points
}
