// Copyright 2013 SoundCloud, Rany Keddo. All rights reserved.  Use of this
// source code is governed by a license that can be found in the LICENSE file.

package bandit

import (
	"encoding/csv"
	"os"
)

var (
	selectLog = csv.NewWriter(os.Stdout)
)

func init() {
	selectLog.Comma = '\t'
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

	return selectLog.WriteAll([][]string{record})
}
