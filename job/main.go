package main

import (
	"flag"
	"github.com/purzelrakete/bandit"
	"log"
	"os"
)

var (
	jobExperimentName = flag.String("experiment-name", "default", "name of the experiment to run")
	jobExperiments    = flag.String("experiments", "experiments.tsv", "experiments tsv filename")
	jobKind           = flag.String("kind", "", "kind ∈ {map,reduce}")
)

func init() {
	flag.Parse()
}

func main() {
	e, err := bandit.NewExperiment(*jobExperiments, *jobExperimentName)
	if err != nil {
		log.Fatalf("could parse experiment: %s", err.Error())
	}

	switch *jobKind {
	case "map":
		bandit.SnapshotMapper(e, os.Stdin, os.Stdout)()
	case "reduce":
		bandit.SnapshotReducer(e, os.Stdin, os.Stdout)()
	case "":
		log.Fatalf("please provide a job kind ∈ {map,reduce}")
	default:
		log.Fatalf("unkown job kind: %s", *jobKind)
	}
}
