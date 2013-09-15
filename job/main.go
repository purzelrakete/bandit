package main

import (
	"flag"
	"github.com/purzelrakete/bandit"
	"log"
	"os"
)

var (
	jobExperimentName = flag.String("experiment-name", "default", "name of experiment")
	jobExperiments    = flag.String("experiments", "experiments.tsv", "experiments tsv")
	jobKind           = flag.String("kind", "", "kind ∈ {map,reduce,poll}")
	jobLogfile        = flag.String("log-file", "bandit-log.txt", "log file to read")
	jobLogPoll        = flag.Duration("log-poll", 1e13, "produce snapshots with this fq")
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
	case "poll":
		if err := simple(e, *jobLogfile, *jobExperimentName+".dsv", *jobLogPoll); err != nil {
			log.Fatalf("could not start polling job: %s", err.Error())
		}
	case "":
		log.Fatalf("please provide a job kind ∈ {map,reduce,poll}")
	default:
		log.Fatalf("unkown job kind: %s", *jobKind)
	}
}
