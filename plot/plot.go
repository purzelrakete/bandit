package main

import (
	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"fmt"
	"github.com/purzelrakete/bandit"
	"image/color"
	"strconv"
	"strings"
)

// graph is labelled plot lines
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

	name := strings.Replace(strings.ToLower(title), " ", "-", -1)
	filename := fmt.Sprintf("bandit-%s.png", name)
	if err := p.Save(8, 8, filename); err != nil {
		return fmt.Errorf(err.Error())
	}

	return nil
}

// xys turns a slice of float64 values into a plotter.XYs
func xys(data []float64) plotter.XYs {
	points := make(plotter.XYs, len(data))
	for i, datum := range data {
		points[i].X = float64(i)
		points[i].Y = datum
	}

	return points
}

// bandits
type bandits []bandit.Bandit

// simulations
type simulations []bandit.Simulation

// arms
type arms []bandit.Arm

// simulate runs a Monte Carlo simulation with given arms and bandits
func simulate(bs bandits, arms arms, sims, horizon int) (simulations, error) {
	ret := simulations{}
	for _, b := range bs {
		s, err := bandit.MonteCarlo(sims, horizon, arms, b)
		if err != nil {
			return simulations{}, fmt.Errorf(err.Error())
		}

		ret = append(ret, s)
	}

	return ret, nil
}

// group ties together n bandits over all summary functions
type group struct {
	name    string
	bandits bandits
}

// summarize summarizes simulations and coverts the to graph
func summarize(sims simulations, summary bandit.Summary) graph {
	lines := make(graph)
	for _, sim := range sims {
		lines[sim.Description] = summary(&sim)
	}

	return lines
}
