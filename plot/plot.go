// Copyright 2013 SoundCloud, Rany Keddo. All rights reserved.  Use of this
// source code is governed by a license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"fmt"
	"github.com/purzelrakete/bandit/sim"
	"image/color"
	"strings"
)

// graph is labelled plot lines
type graph map[string][]float64

func rgb(r, g, b uint8) color.RGBA {
	return color.RGBA{r, g, b, 255}
}

func getColor(i int) color.Color {
	defaultColors := []color.Color{
		rgb(241, 90, 96),
		rgb(122, 195, 106),
		rgb(90, 155, 212),
		rgb(250, 167, 91),
		rgb(158, 103, 171),
		rgb(206, 112, 88),
		rgb(215, 127, 180),
	}

	if i > 0 && i <= len(defaultColors) {
		return defaultColors[i]
	}

	return rgb(0, 0, 0)
}

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
		l.LineStyle.Color = getColor(i)
	}

	if err != nil {
		return fmt.Errorf(err.Error())
	}

	name := strings.Replace(strings.ToLower(title), " ", "-", -1)
	filename := fmt.Sprintf("bandit-%s.svg", name)
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
type bandits []sim.Bandit

// simulations
type simulations []sim.Simulation

// arms
type arms []sim.Arm

// simulate runs a Monte Carlo simulation with given arms and bandits
func simulate(bs bandits, arms arms, sims, horizon int) (simulations, error) {
	ret := simulations{}
	for _, b := range bs {
		s, err := sim.MonteCarlo(sims, horizon, arms, b)
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
func summarize(sims simulations, summary sim.Summary) graph {
	lines := make(graph)
	for _, sim := range sims {
		lines[sim.Description] = summary(&sim)
	}

	return lines
}
