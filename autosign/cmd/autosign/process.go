package main

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-github/v63/github"
	"mutantmonkey.in/code/geschenkerbauer/autosign/internal/fshelpers"
)

func processWorkflowRun(client *github.Client, run *github.WorkflowRun, config Config) error {
	artifacts, _, err := client.Actions.ListWorkflowRunArtifacts(context.Background(), config.GitHub.Owner, config.GitHub.Repo, run.GetID(), nil)
	if err != nil {
		return err
	}
	for _, artifact := range artifacts.Artifacts {
		fmt.Printf("Processing artifact: %s (%d)\n", artifact.GetName(), artifact.GetID())

		url, _, err := client.Actions.DownloadArtifact(context.Background(), config.GitHub.Owner, config.GitHub.Repo, artifact.GetID(), 1)
		if err != nil {
			return fmt.Errorf("could not download artifact: %v", err)
		}

		f, err := os.CreateTemp("", "*.zip")
		if err != nil {
			return fmt.Errorf("could not create temporary file: %v", err)
		}
		defer os.Remove(f.Name())

		resp, err := http.Get(url.String())
		if err != nil {
			return fmt.Errorf("could not download file: %v", err)
		}
		defer resp.Body.Close()

		_, err = io.Copy(f, resp.Body)
		f.Close()
		if err != nil {
			return fmt.Errorf("could not write to destination file: %v", err)
		}

		if err := processArtifact(f.Name(), config); err != nil {
			return fmt.Errorf("could not process artifact: %v", err)
		}
	}

	return nil
}

func processArtifact(filename string, config Config) error {
	repo := fmt.Sprintf("%s/%s", config.GitHub.Owner, config.GitHub.Repo)

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
		if !config.SkipAttestationCheck {
			// TODO: when go-github supports this, do this in pure Go instead
			cmd := exec.Command("gh", "attestation", "verify", destFilepath, "-R", repo)
			cmd.Env = os.Environ()
			cmd.Env = append(cmd.Env, fmt.Sprintf("GH_TOKEN=%s", config.GitHub.AuthToken))
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("error validating attestation: %v", err)
			}
		}

		if err := signPackage(destFilepath, config.Keyring); err != nil {
			return fmt.Errorf("error signing package: %v", err)
		}

		// move package signature to final output directory
		if err := fshelpers.MoveFile(destFilepath+".sig", filepath.Join(config.OutputDir, destFilename+".sig")); err != nil {
			return fmt.Errorf("error moving package signature to destination directory: %v", err)
		}

		// move package to final output directory
		if err := fshelpers.MoveFile(destFilepath, filepath.Join(config.OutputDir, destFilename)); err != nil {
			return fmt.Errorf("error moving package to destination directory: %v", err)
		}

		if !config.SkipRepoAdd {
			// add new packages to repository database
			// TODO: it would be nice if I could do this in pure Go
			cmd := exec.Command("repo-add", config.DbName, destFilename)
			cmd.Dir = config.OutputDir
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("error adding package to repository database: %v", err)
			}
		}
	}

	return nil
}

func ProcessWorkflowRun(config Config, client *github.Client, runID int64) error {
	run, _, err := client.Actions.GetWorkflowRunByID(context.Background(), config.GitHub.Owner, config.GitHub.Repo, runID)
	if err != nil {
		return err
	}

	fmt.Printf("Processing workflow run: %s #%d\n", run.GetName(), run.GetID())
	if err := processWorkflowRun(client, run, config); err != nil {
		return err
	}

	return nil
}

func ProcessWorkflows(config Config, client *github.Client, opts ProcessOptions) error {
	minCreatedAt := time.Now().UTC()
	if opts.TimeWindow != "" {
		duration, err := time.ParseDuration(opts.TimeWindow)
		if err != nil {
			return err
		}
		minCreatedAt = minCreatedAt.Add(-duration)
	}

	if opts.WorkflowName != "" {
		runs, _, err := client.Actions.ListWorkflowRunsByFileName(context.Background(), config.GitHub.Owner, config.GitHub.Repo, opts.WorkflowName, nil)
		if err != nil {
			return err
		}

		for _, run := range runs.WorkflowRuns {
			if opts.MinRunNumber > 0 {
				if run.GetID() < int64(opts.MinRunNumber) {
					break
				}
			} else if run.CreatedAt.GetTime().Compare(minCreatedAt) < 0 {
				break
			}

			fmt.Printf("Processing workflow run: %s #%d\n", run.GetName(), run.GetID())
			if err := processWorkflowRun(client, run, config); err != nil {
				return err
			}
		}
	} else {
		if opts.MinRunNumber > 0 {
			return errors.New("workflow name is required")
		}

		workflows, _, err := client.Actions.ListWorkflows(context.Background(), config.GitHub.Owner, config.GitHub.Repo, nil)
		if err != nil {
			return fmt.Errorf("error listing workflows: %v", err)
		}

		for _, workflow := range workflows.Workflows {
			runs, _, err := client.Actions.ListWorkflowRunsByID(context.Background(), config.GitHub.Owner, config.GitHub.Repo, workflow.GetID(), nil)
			if err != nil {
				log.Fatal(err)
			}

			for _, run := range runs.WorkflowRuns {
				if run.CreatedAt.GetTime().Compare(minCreatedAt) < 0 {
					break
				}

				fmt.Printf("Processing workflow run: %s #%d\n", run.GetName(), run.GetID())
				if err := processWorkflowRun(client, run, config); err != nil {
					return err
				}
			}

		}
	}

	return nil
}
