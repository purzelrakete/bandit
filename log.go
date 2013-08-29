// Copyright 2013 SoundCloud, Rany Keddo. All rights reserved.  Use of this
// source code is governed by a license that can be found in the LICENSE file.

package bandit

import (
	"fmt"
	"log"
	"strings"
)

// LogSelection captures all selected arms. This log can be used in conjunction
// with reward logs to fully rebuild bandits.
func LogSelection(uid string, experiment Experiment, selected Variant) {
	record := []string{
		"BanditSelection",
		uid,
		experiment.Name,
		fmt.Sprintf("%d", selected.Ordinal),
		selected.Tag,
	}

	log.Println(strings.Join(record, " "))
}

// LogReward captures all selected arms. This log can be used in conjunction
// with reward logs to fully rebuild bandits.
func LogReward(uid string, experiment Experiment, selected Variant, reward float64) {
	record := []string{
		"BanditReward",
		uid,
		experiment.Name,
		fmt.Sprintf("%d", selected.Ordinal),
		selected.Tag,
		fmt.Sprintf("%f", reward),
	}

	log.Println(strings.Join(record, " "))
}
