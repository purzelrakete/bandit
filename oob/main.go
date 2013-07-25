package main

import (
	"flag"

	"github.com/purzelrakete/bandit"
	bhttp "github.com/purzelrakete/bandit/http"
	"log"
	"net/http"
)

var (
	campaignsTsv = flag.String("campaigns", "campaigns.tsv", "campaigns tsv filename")
	port         = flag.String("port", ":8080", "http port")
)

func init() {
	flag.Parse()
}

func main() {
	campaigns, err := bandit.ParseCampaigns(*campaignsTsv)
	if err != nil {
		log.Fatalf("could not read campaigns: %s", err.Error())
	}

	for name, campaign := range campaigns {
		arms := len(campaign.Variants)
		b, err := bandit.NewEpsilonGreedy(arms, 0.1)
		if err != nil {
			log.Fatal(err.Error())
		}

		http.HandleFunc("/ab/"+name, bhttp.OOBHandler(b, campaign))
	}

	log.Fatal(http.ListenAndServe(*port, nil))
}
