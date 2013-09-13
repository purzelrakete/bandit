// Copyright 2013 SoundCloud, Rany Keddo. All rights reserved.  Use of this
// source code is governed by a license that can be found in the LICENSE file.

package bandit

import (
	"fmt"
	"time"

	"strings"
)

const (
	banditSelection = "BanditSelection"
	banditReward    = "BanditReward"
)

// SelectionLine captures all selected arms. This log can be used in conjunction
// with reward logs to fully rebuild bandits.
func SelectionLine(experiment Experiment, selected Variant) string {
	record := []string{
		fmt.Sprintf("%d", time.Now().Unix()),
		banditSelection,
		experiment.Name,
		selected.Tag,
	}

	return strings.Join(record, " ")
}

// RewardLine captures all selected arms. This log can be used in conjunction
// with reward logs to fully rebuild bandits.
func RewardLine(experiment Experiment, selected Variant, reward float64) string {
	record := []string{
		fmt.Sprintf("%d", time.Now().Unix()),
		banditReward,
		experiment.Name,
		selected.Tag,
		fmt.Sprintf("%f", reward),
	}

	return strings.Join(record, " ")
}
