package main

import (
	"flag"
	"github.com/purzelrakete/bandit"
	"log"
	"os"
)

var (
	jobExperimentName = flag.String("experiment-name", "default", "name of experiment")
	jobKind           = flag.String("kind", "", "kind ∈ {map,reduce,poll}")
	jobLogfile        = flag.String("log-file", "bandit-log.txt", "log file to read")
	jobLogPoll        = flag.Duration("log-poll", 1e13, "produce snapshots with this fq")
)

func init() {
	flag.Parse()
}

func main() {
	switch *jobKind {
	case "map":
		bandit.SnapshotMapper(*jobExperimentName, os.Stdin, os.Stdout)()
	case "reduce":
		bandit.SnapshotReducer(*jobExperimentName, os.Stdin, os.Stdout)()
	case "poll":
		if err := simple(*jobExperimentName, *jobLogfile, *jobExperimentName+".dsv", *jobLogPoll); err != nil {
			log.Fatalf("could not start polling job: %s", err.Error())
		}
	case "":
		log.Fatalf("please provide a job kind ∈ {map,reduce,poll}")
	default:
		log.Fatalf("unkown job kind: %s", *jobKind)
	}
}
