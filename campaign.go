// Copyright 2013 SoundCloud, Rany Keddo. All rights reserved.  Use of this
// source code is governed by a license that can be found in the LICENSE file.

package bandit

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Campaign is a single campaign. Variants are in ascending ordinal sorting,
// where ordinals are contiguous and start at 1.
type Campaign struct {
	Name     string
	Variants Variants
}

// Variant describes endpoints which are mapped onto bandit arms.
type Variant struct {
	Ordinal int    // 1 indexed arm ordinal
	URL     string // the url associated with this variant, for out of band
	Tag     string // this tag is used throughout the lifecycle of the campaign
}

// Variants is a set of variants sorted by ordinal.
type Variants []Variant

// Len satisfies sort.Interface
func (v Variants) Len() int {
	return len(v)
}

// Less satisfies sort.Interface
func (v Variants) Less(i, j int) bool {
	return v[i].Ordinal < v[j].Ordinal
}

// Swap satisfies sort.Interface
func (v Variants) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

// SelectVariant selects the appropriate variant given it's 1 indexed ordinal
func SelectVariant(c Campaign, ordinal int) (Variant, error) {
	if l := len(c.Variants); ordinal < 0 || ordinal > l {
		return Variant{}, fmt.Errorf("ordinal %d not in [1,%d]", ordinal, l)
	}

	return c.Variants[ordinal-1], nil
}

// GetVariant returns the Campaign and variant pointed to by a string tag.
func GetVariant(t *Tests, tag string) (Campaign, Variant, error) {
	for _, test := range *t {
		for _, variant := range test.Campaign.Variants {
			if variant.Tag == tag {
				return test.Campaign, variant, nil
			}
		}
	}

	return Campaign{}, Variant{}, fmt.Errorf("could not find variant '%s'", tag)
}

// Test is a bandit set up against a campaign.
type Test struct {
	Bandit   Bandit
	Campaign Campaign
}

// Tests maps campaign names to Test setups.
type Tests map[string]Test

// NewTests returns a complete set of campaign, bandit tuples (bandit.Test).
func NewTests(campaignTSV string) (Tests, error) {
	campaigns, err := ParseCampaigns(campaignTSV)
	if err != nil {
		return Tests{}, fmt.Errorf("could not read campaigns: %s", err.Error())
	}

	tests := make(Tests)
	for name, campaign := range campaigns {
		b, err := NewSoftmax(len(campaign.Variants), 0.1)
		if err != nil {
			return Tests{}, fmt.Errorf(err.Error())
		}

		tests[name] = Test{
			Bandit:   b,
			Campaign: campaign,
		}
	}

	return tests, nil
}

// Campaigns is an index of names to campaigns
type Campaigns map[string]Campaign

// ParseCampaigns reads in a tsv file and converts it to a list of campaigns.
func ParseCampaigns(filename string) (Campaigns, error) {
	file, err := os.Open(filename)
	if err != nil {
		return Campaigns{}, fmt.Errorf("need a valid input file: %v", err)
	}

	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = '\t'
	records, err := reader.ReadAll()
	if err != nil {
		return Campaigns{}, fmt.Errorf("could not read tsv: %s ", err)
	}

	// intermediary data structure groups variants
	type campaignVariants map[string]Variants

	variants := make(campaignVariants)
	for i, record := range records {
		if l := len(record); l != 4 {
			return Campaigns{}, fmt.Errorf("record is not %v long: %v", l, record)
		}

		ordinal, err := strconv.Atoi(record[1])
		if err != nil {
			return Campaigns{}, fmt.Errorf("invalid ordinal on line %n: %s", i, err)
		}

		name := record[0]
		if words := strings.Fields(name); len(words) != 1 {
			return Campaigns{}, fmt.Errorf("campaign has whitespace: %s", name)
		}

		tag := record[3]
		if words := strings.Fields(tag); len(words) != 1 {
			return Campaigns{}, fmt.Errorf("tag has whitespace: %s", tag)
		}

		variants[name] = append(variants[name], Variant{
			Ordinal: ordinal,
			URL:     record[2],
			Tag:     tag,
		})
	}

	// sorted campaign variants
	campaigns := make(Campaigns)
	for name, variants := range variants {
		sort.Sort(variants)
		campaigns[name] = Campaign{
			Name:     name,
			Variants: variants,
		}
	}

	// fail if ordinals are non-contiguous or do not start with 1
	for name, variants := range variants {
		for i := 0; i < len(variants); i++ {
			if ord := variants[i].Ordinal; ord != i+1 {
				return Campaigns{}, fmt.Errorf("%s: variant %d noncontiguous", name, ord)
			}
		}
	}

	return campaigns, nil
}
