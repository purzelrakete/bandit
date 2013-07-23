package main

import (
	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"flag"
	"fmt"
	"github.com/purzelrakete/bandit"
	"image/color"
	"log"
	"strconv"
	"strings"
)

// sims maps model parameter such as ε to corresponding simulation results
type sims map[float64]bandit.Simulation

// summary summarizes a Simulation and returns corresponding plot points.
type summary func(s bandit.Simulation) []float64

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

// parseArms converts command line 0.1,0.2 into a slice of floats. Returns
// the the best arm (1 indexed). In the case of equally good best arms the
// last arm is returned.
func parseArms(sμ string) ([]float64, int, error) {
	var μs []float64
	max, imax := 0.0, 0
	for i, s := range strings.Split(*mcMus, ",") {
		μ, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return []float64{}, 0, fmt.Errorf("could not parse float: %s", err.Error())
		}

		if μ < 0 || μ > 1 {
			return []float64{}, 0, fmt.Errorf("μ must be in [0,1]: %.5f", μ)
		}

		if μ > max {
			max = μ
			imax = i
		}

		μs = append(μs, μ)
	}

	return μs, imax + 1, nil
}

var (
	mcSims           = flag.Int("mcSims", 5000, "monte carlo simulations to run")
	mcHorizon        = flag.Int("mcHorizon", 300, "trials per simulation")
	mcMus            = flag.String("mcMus", "0.1,0.3,0.2,0.8", "bernoulli arm μ parameters")
	mcPerformancePng = flag.String("mcPerformancePng", "bandit_performance.png", "performance plot")
	mcAccuracyPng    = flag.String("mcAccuracyPng", "bandit_accuracy.png", "accuracy plot")
	mcCumulativePng  = flag.String("mcCumulativePng", "bandit_cumulative.png", "cumulative plot")
)

func init() {
	flag.Parse()
}

func main() {
	μs, bestArm, err := parseArms(*mcMus)
	if err != nil {
		log.Fatalf(err.Error())
	}

	εs := []float64{0.1, 0.2, 0.3, 0.4, 0.5}
	sims := make(sims)
	for _, ε := range εs {
		var arms []bandit.Arm
		for _, μ := range μs {
			arms = append(arms, bandit.Bernoulli(μ))
		}

		s, err := bandit.MonteCarlo(*mcSims, *mcHorizon, arms, func() (bandit.Bandit, error) {
			return bandit.EpsilonGreedyNew(len(μs), ε)
		})

		if err != nil {
			log.Fatalf(err.Error())
		}

		sims[ε] = s
	}

	title, xLabel, yLabel := "Greedy Accuracy", "Time", "P(selecting best arm)"
	draw(title, xLabel, yLabel, *mcAccuracyPng, sims, func(s bandit.Simulation) []float64 {
		return bandit.Accuracy(s, bestArm)
	})

	title, xLabel, yLabel = "Greedy Performance", "Time", "Reward"
	draw(title, xLabel, yLabel, *mcPerformancePng, sims, func(s bandit.Simulation) []float64 {
		return bandit.Performance(s)
	})

	title, xLabel, yLabel = "Greedy Cumulative Performance", "Time", "Cumulative Reward"
	draw(title, xLabel, yLabel, *mcCumulativePng, sims, func(s bandit.Simulation) []float64 {
		return bandit.Cumulative(s)
	})
}
