// Copyright 2013 SoundCloud, Rany Keddo. All rights reserved.  Use of this
// source code is governed by a license that can be found in the LICENSE file.

package bandit

import (
	"fmt"
	"log"
	"strings"
)

const (
	banditSelection = "BanditSelection"
	banditReward    = "BanditReward"
)

// LogSelection captures all selected arms. This log can be used in conjunction
// with reward logs to fully rebuild bandits.
func LogSelection(experiment Experiment, selected Variant) {
	record := []string{
		banditSelection,
		experiment.Name,
		selected.Tag,
	}

	log.Println(strings.Join(record, " "))
}

// LogReward captures all selected arms. This log can be used in conjunction
// with reward logs to fully rebuild bandits.
func LogReward(experiment Experiment, selected Variant, reward float64) {
	record := []string{
		banditReward,
		experiment.Name,
		selected.Tag,
		fmt.Sprintf("%f", reward),
	}

	log.Println(strings.Join(record, " "))
}
