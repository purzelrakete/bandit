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

// sims maps model parameter such as ε to corresponding simulation results
type sims map[float64]bandit.Sim

// summary summarizes a Sim and returns corresponding plot points.
type summary func(s bandit.Sim) []float64

// xys turns a slice of float64 values into a plotter.XYs
func xys(data []float64) plotter.XYs {
	points := make(plotter.XYs, len(data))
	for i, datum := range data {
		points[i].X = float64(i)
		points[i].Y = datum
	}

	return points
}

// draw is a generic plotter of simulation summaries.
func draw(title, xLabel, yLabel, filename string, sims sims, summary summary) {
	p, err := plot.New()
	if err != nil {
		log.Fatalf(err.Error())
	}

	p.Title.Text = title
	p.X.Label.Text = xLabel
	p.Y.Label.Text = yLabel

	for ε, sim := range sims {
		l, err := plotter.NewLine(xys(summary(sim)))
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

	if err := p.Save(8, 8, filename); err != nil {
		log.Fatalf(err.Error())
	}
}

var (
	mcSims           = flag.Int("mcSims", 5000, "monte carlo simulations to run")
	mcHorizon        = flag.Int("mcHorizon", 300, "trials per simulation")
	mcPerformancePng = flag.String("mcPerformancePng", "performance.png", "performance plot")
	mcAccuracyPng    = flag.String("mcAccuracyPng", "accuracy.png", "accuracy plot")
	mcCumulativePng  = flag.String("mcCumulativePng", "cumulative.png", "cumulative plot")
)

func init() {
	flag.Parse()
}

func main() {
	εs := []float64{0.1, 0.2, 0.3, 0.4, 0.5}
	μs := []float64{0.1, 0.3, 0.2, 0.8}
	sims := make(sims)

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

	title, xLabel, yLabel := "Greedy Accuracy", "Time", "P(selecting best arm)"
	draw(title, xLabel, yLabel, *mcAccuracyPng, sims, func(s bandit.Sim) []float64 {
		return bandit.Accuracy(s, 4)
	})

	title, xLabel, yLabel = "Greedy Performance", "Time", "Reward"
	draw(title, xLabel, yLabel, *mcPerformancePng, sims, func(s bandit.Sim) []float64 {
		return bandit.Performance(s)
	})

	title, xLabel, yLabel = "Greedy Cumulative Performance", "Time", "Cumulative Reward"
	draw(title, xLabel, yLabel, *mcCumulativePng, sims, func(s bandit.Sim) []float64 {
		return bandit.Cumulative(s)
	})
}
