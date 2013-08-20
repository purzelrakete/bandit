// Copyright 2013 SoundCloud, Rany Keddo. All rights reserved.  Use of this
// source code is governed by a license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"github.com/bmizerany/pat"
	"github.com/purzelrakete/bandit"
	bhttp "github.com/purzelrakete/bandit/http"
	"log"
	"net/http"
)

var (
	exCampaigns = flag.String("campaigns", "campaigns.tsv", "campaigns tsv filename")
	exBind      = flag.String("bind", ":8080", "interface and port to bind to")
)

func init() {
	flag.Parse()
}

// index serves the index relative to the project root.
func index(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	http.ServeFile(w, r, "example/index.html")
}

// widget a jsonp response to render the required widget.
func widget(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/javascript")

	shape := r.URL.Query().Get("shape")
	callback := r.URL.Query().Get("callback")
	js := fmt.Sprintf("%s({ shape: '%s' });", callback, shape)
	w.Write([]byte(js))
}

func main() {
	tests, err := bandit.NewTests(*exCampaigns)
	if err != nil {
		log.Fatalf("could not construct campaigns: %s", err.Error())
	}

	// routes
	m := pat.New()
	m.Get("/", http.HandlerFunc(index))
	m.Get("/widget", http.HandlerFunc(widget))
	m.Get("/select/:campaign", http.HandlerFunc(bhttp.SelectionHandler(tests)))
	m.Get("/feedback", http.HandlerFunc(bhttp.LogRewardHandler(tests)))
	http.Handle("/", m)

	// serve
	log.Fatal(http.ListenAndServe(*exBind, nil))
}
