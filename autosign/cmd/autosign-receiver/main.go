package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	IncomingDir string
	RepoDir     string
	DbName      string
	GitHub      GitHubConfig
	Receiver    ReceiverConfig
}

type GitHubConfig struct {
	Owner     string
	Repo      string
	AuthToken string
}

type ReceiverConfig struct {
	ListenAddr  string `default:":8080"`
	BearerToken string
}

func main() {
	var configPath string
	var daemonMode bool
	flag.StringVar(&configPath, "config", "", "Path to TOML configuration file")
	flag.BoolVar(&daemonMode, "d", false, "Run in daemon mode and listen for webhook calls")
	flag.Parse()

	if configPath == "" {
		// TODO: this should use a default path if one is not specified
		log.Fatal("The -config option is required.")
	}

	config := Config{}

	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		log.Fatalf("Error reading config: %v", err)
	}

	if daemonMode {
		var processMutex sync.Mutex

		s := http.Server{
			Addr:           config.Receiver.ListenAddr,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}

		http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
				return
			}

			if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %s", config.Receiver.BearerToken) {
				http.Error(w, "403 forbidden", http.StatusForbidden)
				return
			}

			go func() {
				processMutex.Lock()

				err := ProcessIncoming(config)
				if err != nil {
					log.Print(err)
				}

				processMutex.Unlock()
			}()

			fmt.Fprintf(w, "ok\n")
		})

		log.Fatal(s.ListenAndServe())
	} else {
		err := ProcessIncoming(config)
		if err != nil {
			log.Fatal(err)
		}
	}
}
