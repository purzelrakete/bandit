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
	"time"
)

var (
	apiExperiments = flag.String("experiments", "experiments.tsv", "experiments tsv filename")
	apiBind        = flag.String("port", ":8080", "interface / port to bind to")
	apiSnapshot    = flag.String("snapshot", "snapshot.dsv", "campaign snapshot file")
)

func init() {
	flag.Parse()
}

func main() {
	t, err := bandit.NewDelayedTrials(*apiExperiments, *apiSnapshot, 1*time.Minute)
	if err != nil {
		log.Fatalf("could not construct experiments: %s", err.Error())
	}

	m := pat.New()
	m.Get("/select/:experiment", http.HandlerFunc(bhttp.SelectionHandler(t)))
	http.Handle("/", m)

	// serve
	log.Fatal(http.ListenAndServe(*apiBind, nil))
}
