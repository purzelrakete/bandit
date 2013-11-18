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
	exExperiments = flag.String("experiments", "experiments.json", "experiments json filename")
	exBind        = flag.String("bind", ":8080", "interface and port to bind to")
	exPinTTL      = flag.Duration("pin-ttl", 0, "ttl life of a pinned variation")
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
	e, err := bandit.NewExperiments(bandit.NewFileOpener(*exExperiments))
	if err != nil {
		log.Fatalf("could not construct experiments: %s", err.Error())
	}

	// routes
	mux := pat.New()
	mux.Get("/es/:name", bhttp.SelectionHandler(e, *exPinTTL))
	mux.Get("/widget", http.HandlerFunc(widget))
	mux.Get("/feedback", bhttp.LogRewardHandler(e))
	mux.Get("/", http.HandlerFunc(index))
	http.Handle("/", mux)

	// serve
	log.Fatal(http.ListenAndServe(*exBind, nil))
}
