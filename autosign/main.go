package main

import (
	"flag"
	"log"
)

type Config struct {
	Owner                string
	Repo                 string
	OutputDir            string
	DbName               string
	Keyring              string
	SkipRepoAdd          bool
	SkipAttestationCheck bool
	AuthToken            string
}

type ProcessOptions struct {
	ConfigPath   string
	TimeWindow   string
	MinRunNumber int
	WorkflowName string
}

func main() {
	opts := ProcessOptions{}
	flag.StringVar(&opts.ConfigPath, "config", "", "Path to TOML configuration file")
	flag.StringVar(&opts.TimeWindow, "since", "4h", "Use workflow runs within this time window")
	flag.IntVar(&opts.MinRunNumber, "run", 0, "Use workflow runs starting with this run number")
	flag.StringVar(&opts.WorkflowName, "workflow", "", "Basename of workflow file")
	flag.Parse()

	if opts.ConfigPath == "" {
		// TODO: this should use a default path if one is not specified
		log.Fatal("The -config option is required.")
	}

	err := ProcessWorkflows(opts)
	if err != nil {
		log.Fatal(err)
	}
}
