package main

import (
	"archive/zip"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/google/go-github/v63/github"
)

type Config struct {
	Owner       string
	Repo        string
	OutputDir   string
	DbName      string
	Keyring     string
	SkipRepoAdd bool
}

func moveFile(oldpath string, newpath string) error {
	srcFile, err := os.Open(oldpath)
	if err != nil {
		return fmt.Errorf("Could not open source file: %s", err)
	}

	destFile, err := os.Create(newpath)
	if err != nil {
		return fmt.Errorf("Could not open destination file: %s", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	srcFile.Close()
	if err != nil {
		return fmt.Errorf("Could not write to destination file: %s", err)
	}

	// Remove original file
	if err := os.Remove(oldpath); err != nil {
		return fmt.Errorf("Could not remove original file: %s", err)
	}

	return nil
}

func processWorkflowRun(client *github.Client, run *github.WorkflowRun, config Config) error {
	artifacts, _, err := client.Actions.ListWorkflowRunArtifacts(context.Background(), config.Owner, config.Repo, run.GetID(), nil)
	if err != nil {
		return err
	}
	for _, artifact := range artifacts.Artifacts {
		fmt.Printf("Processing artifact: %s (%d)\n", artifact.GetName(), artifact.GetID())

		url, _, err := client.Actions.DownloadArtifact(context.Background(), config.Owner, config.Repo, artifact.GetID(), 1)
		if err != nil {
			return fmt.Errorf("Could not download artifact: %s", err)
		}

		f, err := os.CreateTemp("", "*.zip")
		if err != nil {
			return fmt.Errorf("Could not create temporary file: %s", err)
		}
		defer os.Remove(f.Name())

		resp, err := http.Get(url.String())
		if err != nil {
			return fmt.Errorf("Could not download file: %s", err)
		}
		defer resp.Body.Close()

		_, err = io.Copy(f, resp.Body)
		f.Close()
		if err != nil {
			return fmt.Errorf("Could not write to destination file: %s", err)
		}

		if err := processArtifact(f.Name(), config); err != nil {
			return fmt.Errorf("Could not process artifact: %s", err)
		}
	}

	return nil
}

func processArtifact(filename string, config Config) error {
	repo := fmt.Sprintf("%s/%s", config.Owner, config.Repo)

	// create temporary destination directory
	dir, err := os.MkdirTemp("", "geschenkerbauer")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	r, err := zip.OpenReader(filename)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// skip directories, we only want package files
		if f.FileInfo().IsDir() {
			continue
		}

		// skip files that don't end in .pkg.tar.zst
		if !strings.HasSuffix(f.Name, ".pkg.tar.zst") {
			log.Printf("Warning: skipping %q because the filename doesn't end in .pkg.tar.zst", f.Name)
			continue
		}

		destFilename := filepath.Base(filepath.Clean(f.Name))

		// GitHub Actions forbids : in filenames, so the build workflow
		// replaces them before creating the ZIP. We need to replace
		// them back in the filename before creating the file.
		destFilename = strings.ReplaceAll(destFilename, "__3A__", ":")

		destFilepath := filepath.Join(dir, destFilename)
		fmt.Printf("%s\n", destFilepath)

		// skip packages that already exist in the output directory
		if _, err := os.Stat(filepath.Join(config.OutputDir, destFilename)); err == nil {
			log.Printf("Warning: skipping %q because it already exists in the output directory", destFilename)
			continue
		}

		df, err := os.Create(destFilepath)
		if err != nil {
			return err
		}

		zf, err := f.Open()
		if err != nil {
			return err
		}

		if _, err := io.Copy(df, zf); err != nil {
			return err
		}

		if err := zf.Close(); err != nil {
			return err
		}

		if err := df.Close(); err != nil {
			return err
		}

		// verify attestation
		// TODO: when go-github supports this, do this in pure Go instead
		cmd := exec.Command("gh", "attestation", "verify", destFilepath, "-R", repo)
		if err := cmd.Run(); err != nil {
			return err
		}

		if err := signPackage(destFilepath, config.Keyring); err != nil {
			return err
		}

		// move package signature to final output directory
		if err := moveFile(destFilepath+".sig", filepath.Join(config.OutputDir, destFilename+".sig")); err != nil {
			return err
		}

		// move package to final output directory
		if err := moveFile(destFilepath, filepath.Join(config.OutputDir, destFilename)); err != nil {
			return err
		}

		if !config.SkipRepoAdd {
			// add new packages to repository database
			// TODO: it would be nice if I could do this in pure Go
			cmd = exec.Command("repo-add", config.DbName, destFilename)
			cmd.Dir = config.OutputDir
			if err := cmd.Run(); err != nil {
				return err
			}
		}
	}

	return nil
}

func main() {
	var configPath string
	var timeWindow string
	var minRunNumber int
	var workflowName string
	flag.StringVar(&configPath, "config", "", "Path to TOML configuration file")
	flag.StringVar(&timeWindow, "since", "4h", "Use workflow runs within this time window")
	flag.IntVar(&minRunNumber, "run", 0, "Use workflow runs starting with this run number")
	flag.StringVar(&workflowName, "workflow", "", "Basename of workflow file")
	flag.Parse()

	if configPath == "" {
		// TODO: this should use a default path if one is not specified
		log.Fatal("The -config option is required.")
	}

	// TODO: parse this from a config file
	cmd := exec.Command("gh", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	token := strings.TrimSpace(string(output))

	config := Config{
		SkipRepoAdd: true,
	}

	if _, err = toml.DecodeFile(configPath, &config); err != nil {
		log.Fatal(err)
	}

	minCreatedAt := time.Now().UTC()
	if timeWindow != "" {
		duration, err := time.ParseDuration(timeWindow)
		if err != nil {
			log.Fatal(err)
		}
		minCreatedAt = minCreatedAt.Add(-duration)
	}

	client := github.NewClient(nil).WithAuthToken(token)

	if workflowName != "" {
		runs, _, err := client.Actions.ListWorkflowRunsByFileName(context.Background(), config.Owner, config.Repo, workflowName, nil)
		if err != nil {
			log.Fatal(err)
		}

		for _, run := range runs.WorkflowRuns {
			if minRunNumber > 0 {
				if run.GetID() < int64(minRunNumber) {
					break
				}
			} else if run.CreatedAt.GetTime().Compare(minCreatedAt) < 0 {
				break
			}

			fmt.Printf("Processing workflow run: %s #%d\n", run.GetName(), run.GetID())
			if err := processWorkflowRun(client, run, config); err != nil {
				log.Fatal(err)
			}
		}
	} else {
		if minRunNumber > 0 {
			log.Fatal("A workflow name must be specified with the -workflow flag when using the -run flag.")
		}

		workflows, _, err := client.Actions.ListWorkflows(context.Background(), config.Owner, config.Repo, nil)
		if err != nil {
			log.Fatal(err)
		}

		for _, workflow := range workflows.Workflows {
			runs, _, err := client.Actions.ListWorkflowRunsByID(context.Background(), config.Owner, config.Repo, workflow.GetID(), nil)
			if err != nil {
				log.Fatal(err)
			}

			for _, run := range runs.WorkflowRuns {
				if run.CreatedAt.GetTime().Compare(minCreatedAt) < 0 {
					break
				}

				fmt.Printf("Processing workflow run: %s #%d\n", run.GetName(), run.GetID())
				if err := processWorkflowRun(client, run, config); err != nil {
					log.Fatal(err)
				}
			}

		}
	}
}
