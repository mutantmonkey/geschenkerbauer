package main

import (
	"encoding/json"
	"flag"
	"log"

	"git.sr.ht/~emersion/go-smee"
	"github.com/BurntSushi/toml"
	"github.com/google/go-github/v66/github"
	"mutantmonkey.in/code/geschenkerbauer/autosign/internal/configs"
)

type Config struct {
	OutputDir            string
	DbName               string
	Keyring              string
	SkipRepoAdd          bool
	SkipAttestationCheck bool
	SmeeProxyURL         string
	GitHub               GitHubConfig
	Webhook              WebhookConfig
	S3                   configs.S3Config
}

type GitHubConfig struct {
	Owner         string
	Repo          string
	AuthToken     string
	WebhookSecret string
}

type WebhookConfig struct {
	URL         string
	BearerToken string
}

type ProcessOptions struct {
	ConfigPath   string
	TimeWindow   string
	MinRunNumber int
	WorkflowName string
}

func main() {
	var configPath string
	var listenForRuns bool
	opts := ProcessOptions{}
	flag.StringVar(&configPath, "config", "", "Path to TOML configuration file")
	flag.StringVar(&opts.TimeWindow, "since", "4h", "Use workflow runs within this time window")
	flag.IntVar(&opts.MinRunNumber, "run", 0, "Use workflow runs starting with this run number")
	flag.StringVar(&opts.WorkflowName, "workflow", "", "Basename of workflow file")
	flag.BoolVar(&listenForRuns, "listen", false, "Listen for workflow runs (using webhooks via smee.io)")
	flag.Parse()

	if configPath == "" {
		// TODO: this should use a default path if one is not specified
		log.Fatal("The -config option is required.")
	}

	config := Config{
		SkipRepoAdd: true,
	}

	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		log.Fatalf("Error reading config: %v", err)
	}

	client := github.NewClient(nil).WithAuthToken(config.GitHub.AuthToken)

	if listenForRuns {
		if config.SmeeProxyURL == "" {
			log.Fatal("The SmeeProxyURL option must be set in the configuration when using -listen.")
		}

		ch, err := smee.CreateChannel(config.SmeeProxyURL)
		if err != nil {
			log.Fatal(err)
		}

		for {
			wh, err := ch.ReadWebHook()
			if err != nil {
				log.Fatalf("failed to read webhook: %v", err)
			}

			if wh.Header["x-github-event"] == "workflow_run" {
				err := github.ValidateSignature(wh.Header["x-hub-signature"], wh.Body, []byte(config.GitHub.WebhookSecret))
				if err != nil {
					log.Printf("failed to verify signature: %v", err)
					continue
				}

				event := &github.WorkflowRunEvent{}
				if err := json.Unmarshal(wh.Body, event); err != nil {
					log.Printf("error unmarshaling webhook body: %v", err)
					continue
				}

				if event.GetAction() == "completed" && event.Repo.Owner.GetLogin() == config.GitHub.Owner && event.Repo.GetName() == config.GitHub.Repo {
					err = ProcessWorkflowRun(config, client, *event.WorkflowRun.ID)
					if err != nil {
						log.Print(err)
					}

					err = postProcess(config)
					if err != nil {
						log.Print(err)
					}
				}
			}
		}
	} else {
		err := ProcessWorkflows(config, client, opts)
		if err != nil {
			log.Fatal(err)
		}

		err = postProcess(config)
		if err != nil {
			log.Print(err)
		}
	}
}
