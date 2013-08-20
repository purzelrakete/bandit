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
	oobCampaigns = flag.String("campaigns", "campaigns.tsv", "campaigns tsv filename")
	oobBind      = flag.String("port", ":8080", "interface / port to bind to")
)

func init() {
	flag.Parse()
}

func main() {
	tests, err := bandit.NewTests(*oobCampaigns)
	if err != nil {
		log.Fatalf("could not construct campaigns: %s", err.Error())
	}

	// handlers
	m := pat.New()
	m.Get("/test/:campaign", http.HandlerFunc(bhttp.SelectionHandler(tests)))
	http.Handle("/", m)

	// serve
	log.Fatal(http.ListenAndServe(*oobBind, nil))
}
