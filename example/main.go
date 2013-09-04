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
	exExperiments = flag.String("experiments", "experiments.tsv", "experiments tsv filename")
	exBind        = flag.String("bind", ":8080", "interface and port to bind to")
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
	tests, err := bandit.NewTests(*exExperiments, func(arms int) (bandit.Bandit, error) {
		return bandit.NewSoftmax(arms, 0.1)
	})

	if err != nil {
		log.Fatalf("could not construct experiments: %s", err.Error())
	}

	// routes
	mux := pat.New()
	mux.Get("/select/:experiment", bhttp.SelectionHandler(tests))
	mux.Get("/widget", http.HandlerFunc(widget))
	mux.Get("/feedback", bhttp.LogRewardHandler(tests))
	mux.Get("/", http.HandlerFunc(index))
	http.Handle("/", mux)

	// serve
	log.Fatal(http.ListenAndServe(*exBind, nil))
}
