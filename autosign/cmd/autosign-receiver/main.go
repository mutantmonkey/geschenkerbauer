package main

import (
	"flag"
	"log"

	"github.com/BurntSushi/toml"
)

type Config struct {
	IncomingDir string
	RepoDir     string
	DbName      string
	ListenAddr  string
	GitHub      GitHubConfig
}

type GitHubConfig struct {
	Owner       string
	Repo        string
	AuthToken   string
}

func main() {
	var configPath string
	// TODO: listen port
	flag.StringVar(&configPath, "config", "", "Path to TOML configuration file")
	flag.Parse()

	if configPath == "" {
		// TODO: this should use a default path if one is not specified
		log.Fatal("The -config option is required.")
	}

	config := Config{}

	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		log.Fatalf("Error reading config: %v", err)
	}

	// TODO: use gokr-rsync to copy files over?

	err := ProcessIncoming(config)
	if err != nil {
		log.Fatal(err)
	}
}
