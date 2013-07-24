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

// simulations maps model parameter to corresponding simulation results
type simulations map[float64]bandit.Simulation

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

// plotLine represents labelled plot lines
type plotLines map[string][]float64

// draw is a generic plotter of labelled lines.
func draw(lines plotLines, title, xLabel, yLabel string) error {
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
	mcMus     = flag.String("mcMus", "0.1,0.3,0.2,0.8", "bernoulli arm μ parameters")
)

func init() {
	flag.Parse()
}

func main() {
	μs, bestArms, err := parseArms(*mcMus)
	if err != nil {
		log.Fatalf(err.Error())
	}

	var arms []bandit.Arm
	for _, μ := range μs {
		arms = append(arms, bandit.Bernoulli(μ))
	}

	lines := make(plotLines)
	for _, ε := range []float64{0.1, 0.2, 0.3, 0.4, 0.5} {
		s, err := bandit.MonteCarlo(*mcSims, *mcHorizon, arms, func() (bandit.Bandit, error) {
			return bandit.EpsilonGreedyNew(len(μs), ε)
		})

		if err != nil {
			log.Fatalf(err.Error())
		}

		lines[fmt.Sprintf("EpsilonGreedy(%.2f)", ε)] = bandit.Accuracy(s, bestArms)
	}

	draw(lines, "Accuracy", "Time", "P(selecting best arm)")

	lines = make(plotLines)
	for _, ε := range []float64{0.1, 0.2, 0.3, 0.4, 0.5} {
		s, err := bandit.MonteCarlo(*mcSims, *mcHorizon, arms, func() (bandit.Bandit, error) {
			return bandit.EpsilonGreedyNew(len(μs), ε)
		})

		if err != nil {
			log.Fatalf(err.Error())
		}

		lines[fmt.Sprintf("EpsilonGreedy(%.2f)", ε)] = bandit.Performance(s)
	}

	draw(lines, "Performance", "Time", "P(selecting best arm)")

	lines = make(plotLines)
	for _, ε := range []float64{0.1, 0.2, 0.3, 0.4, 0.5} {
		s, err := bandit.MonteCarlo(*mcSims, *mcHorizon, arms, func() (bandit.Bandit, error) {
			return bandit.EpsilonGreedyNew(len(μs), ε)
		})

		if err != nil {
			log.Fatalf(err.Error())
		}

		lines[fmt.Sprintf("EpsilonGreedy(%.2f)", ε)] = bandit.Cumulative(s)
	}

	draw(lines, "Cumulative", "Time", "P(selecting best arm)")
}
