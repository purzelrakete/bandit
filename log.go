// Copyright 2013 SoundCloud, Rany Keddo. All rights reserved.  Use of this
// source code is governed by a license that can be found in the LICENSE file.

package bandit

import (
	"encoding/csv"
	"fmt"
	"os"
)

var (
	selectWriter = csv.NewWriter(os.Stdout)
	rewardWriter = csv.NewWriter(os.Stdout)
)

func init() {
	selectWriter.Comma = '\t'
}

// LogSelection captures all selected arms. This log can be used in conjunction
// with reward logs to fully rebuild bandits.
func LogSelection(uid string, campaign Campaign, selected Variant) error {
	record := []string{
		uid,
		campaign.Name,
		string(selected.Ordinal),
		selected.Tag,
	}

	return selectWriter.WriteAll([][]string{record})
}

// LogReward captures all selected arms. This log can be used in conjunction
// with reward logs to fully rebuild bandits.
func LogReward(uid string, campaign Campaign, selected Variant, reward float64) error {
	record := []string{
		uid,
		campaign.Name,
		string(selected.Ordinal),
		selected.Tag,
		fmt.Sprintf("%f", reward),
	}

	return rewardWriter.WriteAll([][]string{record})
}
