package main

import (
	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"code.google.com/p/plotinum/vg"
	"flag"
	"fmt"
	"github.com/purzelrakete/bandit"
	"image/color"
	"log"
)

var (
	mcSims    = flag.Int("mcSims", 5000, "monte carlo simulations to run")
	mcHorizon = flag.Int("mcHorizon", 250, "trials per simulation")
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
		l, err := plotter.NewLine(accuracy(sim))
		if err != nil {
			log.Fatalf(err.Error())
		}

		l.LineStyle.Width = vg.Points(1)
		l.LineStyle.Color = color.Gray{uint8(255 * ε)}
		p.Legend.Add(fmt.Sprintf("%.2f", ε), l)
		p.Add(l)
	}

	if err != nil {
		log.Fatalf(err.Error())
	}

	if err := p.Save(8, 8, *mcPerfPng); err != nil {
		log.Fatalf(err.Error())
	}
}

func accuracy(s bandit.Sim) plotter.XYs {
	pts := make(plotter.XYs, *mcHorizon)
	for trial := 0; trial < *mcHorizon; trial++ {
		accum, count := 0.0, 0
		for sim := range pts {
			i := *mcHorizon*sim + trial
			if s.Trial[i] != trial+1 {
				panic("impossible trial access")
			}

			accum = accum + s.Reward[i]
			count = count + 1
		}

		pts[trial].X = float64(trial)
		pts[trial].Y = accum / float64(count)
	}

	return pts
}
