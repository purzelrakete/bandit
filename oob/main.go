package main

import (
	"flag"
	"github.com/bmizerany/pat"
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

	tests := make(bandit.Tests)
	for name, campaign := range campaigns {
		b, err := bandit.NewEpsilonGreedy(len(campaign.Variants), 0.1)
		if err != nil {
			log.Fatal(err.Error())
		}

		tests[name] = bandit.Test{
			Bandit:   b,
			Campaign: campaign,
		}
	}

	m := pat.New()
	m.Get("/test/:campaign", http.HandlerFunc(bhttp.OOBHandler(tests)))
	http.Handle("/", m)
	log.Fatal(http.ListenAndServe(*port, nil))
}
