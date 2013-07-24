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

// xys turns a slice of float64 values into a plotter.XYs
func xys(data []float64) plotter.XYs {
	points := make(plotter.XYs, len(data))
	for i, datum := range data {
		points[i].X = float64(i)
		points[i].Y = datum
	}

	return points
}

// plotLine represents labelled plot lines
type graph map[string][]float64

// draw is a generic plotter of labelled lines.
func draw(lines graph, title, xLabel, yLabel string) error {
	p, err := plot.New()
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	p.Title.Text = title
	p.X.Label.Text = xLabel
	p.Y.Label.Text = yLabel

	i := 0
	for legend, data := range lines {
		i = i + 1
		l, err := plotter.NewLine(xys(data))
		if err != nil {
			return fmt.Errorf(err.Error())
		}

		p.Add(l)
		p.Legend.Add(legend, l)
		l.LineStyle.Color = color.Gray{uint8(48 * float64(i))}
	}

	if err != nil {
		return fmt.Errorf(err.Error())
	}

	filename := fmt.Sprintf("bandit_%s.png", strings.ToLower(title))
	if err := p.Save(8, 8, filename); err != nil {
		return fmt.Errorf(err.Error())
	}

	return nil
}

// banditNew constructs parameterized BanditNew functions
type banditNew func(x float64) bandit.BanditNew

// simulations
type simulations []bandit.Simulation

// arms
type arms []bandit.Arm

// simulate runs a Monte Carlo simulation with given arms and bandits
func simulate(b banditNew, arms arms, sims, horizon int) (simulations, error) {
	ret := simulations{}
	for _, x := range []float64{0.1, 0.2, 0.3, 0.4, 0.5} {
		s, err := bandit.MonteCarlo(sims, horizon, arms, b(x))
		if err != nil {
			return simulations{}, fmt.Errorf(err.Error())
		}

		ret = append(ret, s)
	}

	return ret, nil
}

// summarize summarizes simulations and coverts the to graph
func summarize(sims simulations, summary bandit.Summary) graph {
	lines := make(graph)
	for _, sim := range sims {
		lines[sim.Description] = summary(sim)
	}

	return lines
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
			return []float64{}, []int{}, fmt.Errorf("NaN: %s", err.Error())
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
		log.Fatalf(err.Error())
	}

	arms := arms{}
	for _, μ := range μs {
		arms = append(arms, bandit.Bernoulli(μ))
	}

	bandits := []banditNew{
		func(ε float64) bandit.BanditNew {
			return func() (bandit.Bandit, error) {
				return bandit.EpsilonGreedyNew(len(μs), ε)
			}
		},
	}

	for _, b := range bandits {
		s, err := simulate(b, arms, *mcSims, *mcHorizon)
		if err != nil {
			log.Fatalf(err.Error())
		}

		graph := summarize(s, bandit.Accuracy(bestArms))
		draw(graph, "Accuracy", "Time", "P(selecting best arm)")

		s, err = simulate(b, arms, *mcSims, *mcHorizon)
		if err != nil {
			log.Fatalf(err.Error())
		}

		graph = summarize(s, bandit.Performance)
		draw(graph, "Performance", "Time", "P(selecting best arm)")

		s, err = simulate(b, arms, *mcSims, *mcHorizon)
		if err != nil {
			log.Fatalf(err.Error())
		}

		graph = summarize(s, bandit.Cumulative)
		draw(graph, "Cumulative", "Time", "P(selecting best arm)")
	}
}
