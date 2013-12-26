// Package main contains bandit-job, which takes as input a log of selects and
// rewards of the following format:
//
// 1379257984 BanditSelection shape-20130822:1:8932478932
// 1379257987 BanditReward shape-20130822:1:8932478932 0.000000
//
// Fields are interpreted as follows:
//
// (logline-timestamp, kind, tag, reward)
//
// Tags are interpreted as:
//
// experiment-name:variation-ordinal:pinning-time
//
package main

import (
	"flag"
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
	stats := NewStatistics(*jobExperimentName)

	switch *jobKind {
	case "map":
		mapper(stats, os.Stdin, os.Stdout)()
	case "reduce":
		reducer(stats, os.Stdin, os.Stdout)()
	case "collect":
		collector(stats, os.Stdin, os.Stdout)()
	case "poll":
		if err := simple(stats, *jobLogfile, *jobLogPoll); err != nil {
			log.Fatalf("could not start polling job: %s", err.Error())
		}
	case "":
		log.Fatalf("please provide a job kind ∈ {map,reduce,poll}")
	default:
		log.Fatalf("unkown job kind: %s", *jobKind)
	}
}
