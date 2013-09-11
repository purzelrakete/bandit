// Copyright 2013 SoundCloud, Rany Keddo. All rights reserved.  Use of this
// source code is governed by a license that can be found in the LICENSE file.

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
	apiExperiments = flag.String("experiments", "experiments.tsv", "experiments tsv filename")
	apiBind        = flag.String("port", ":8080", "interface / port to bind to")
	apiSnapshot    = flag.String("snapshot", "snapshot.dsv", "campaign snapshot file")
	apiSnaphotPoll = flag.Duration("snapshot-poll", 1000000000, "time before new snapshot is loaded")
	apiPinTTL      = flag.Duration("pin-ttl", 0, "ttl life of a pinned variant")
)

func init() {
	flag.Parse()
}

func main() {
	es, err := bandit.NewExperiments(*apiExperiments)
	if err != nil {
		log.Fatalf("could not construct experiments: %s", err.Error())
	}

	opener := bandit.NewFileOpener(*apiSnapshot)
	if err := es.InitDelayedBandit(opener, *apiSnaphotPoll); err != nil {
		log.Fatalf("could initialize bandits: %s", err.Error())
	}

	m := pat.New()
	m.Get("/experiments/:name", http.HandlerFunc(bhttp.SelectionHandler(es, *apiPinTTL)))
	http.Handle("/", m)

	// serve
	log.Fatal(http.ListenAndServe(*apiBind, nil))
}
