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
	apiExperiments = flag.String("experiments", "experiments.json", "local file or http endpoint")
	apiBind        = flag.String("port", ":8080", "interface / port to bind to")
	apiPinTTL      = flag.Duration("pin-ttl", 0, "ttl life of a pinned variation")
)

func init() {
	flag.Parse()
}

func main() {
	es, err := bandit.NewExperiments(bandit.NewOpener(*apiExperiments))
	if err != nil {
		log.Fatalf("could not initialize experiments: %s", err.Error())
	}

	m := pat.New()
	m.Get("/experiments/:name", http.HandlerFunc(bhttp.SelectionHandler(es, *apiPinTTL)))
	http.Handle("/", m)

	// serve
	log.Fatal(http.ListenAndServe(*apiBind, nil))
}
