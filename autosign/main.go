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

	"github.com/google/go-github/v62/github"
)

func processArtifact(filename string, repo string) error {
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

		// GitHub Actions forbids : in filenames, so the build action
		// replaces them before creating the ZIP. We need to replace
		// them back in the filename before creating the file.
		destFilename = strings.ReplaceAll(destFilename, "__3A__", ":")

		destFilepath := filepath.Join(dir, destFilename)
		fmt.Printf("%s\n", destFilepath)

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

		// GitHub Actions forbids : in filenames, so the build action
		// replaces them before creating the ZIP. Now that we have the
		// file, rename it back.
		/*newFilename := strings.ReplaceAll(destFilename, "__3A__", ":")
		newFilepath := filepath.Join(dir, newFilename)
		if err := os.Rename(destFilepath, newFilepath); err != nil {
			return err
		}*/

		// TODO: sign - can this be done in pure Go?
		// 	* golang.org/x/crypto/openpgp
		/*if err := signPackage(destFilepath); err != nil {
			return err
		}*/

		// TODO: copy output files (including signatures) to repository
		//	* could consider using os.Rename

		// TODO: call repo-add
		// 	* it's just a shell script, so may want to rewrite in Go
	}

	/*unpackCmd := exec.Command("bash", "../unpack_and_sign.sh", f.Name())
	if err := unpackCmd.Run(); err != nil {
		return err
	}*/

	return nil
}

func main() {
	var timeWindow string
	var minRunNumber int
	var workflowName string
	flag.StringVar(&timeWindow, "since", "4h", "Use workflow runs within this time window")
	flag.IntVar(&minRunNumber, "run", 0, "Use workflow runs starting with this run number")
	flag.StringVar(&workflowName, "workflow", "build_updated_packages.yml", "Basename of workflow file")
	flag.Parse()

	// TODO: parse this from a config file
	cmd := exec.Command("gh", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	token := strings.TrimSpace(string(output))

	// TODO: accept these as flags? or parse from config?
	owner := "mutantmonkey"
	repo := "aur"

	client := github.NewClient(nil).WithAuthToken(token)

	// TODO: should I also support running without specifying a workflow? or is that not needed?
	runs, _, err := client.Actions.ListWorkflowRunsByFileName(context.Background(), owner, repo, workflowName, nil)
	if err != nil {
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

	for _, run := range runs.WorkflowRuns {
		if minRunNumber > 0 {
			if run.GetID() < int64(minRunNumber) {
				break
			}
		} else if run.CreatedAt.GetTime().Compare(minCreatedAt) < 0 {
			break
		}

		fmt.Printf("Processing workflow run: %s #%d\n", run.GetName(), run.GetID())

		artifacts, _, err := client.Actions.ListWorkflowRunArtifacts(context.Background(), owner, repo, run.GetID(), nil)
		if err != nil {
			log.Fatal(err)
		}
		for _, artifact := range artifacts.Artifacts {
			fmt.Printf("Processing artifact: %s (%d)\n", artifact.GetName(), artifact.GetID())

			url, _, err := client.Actions.DownloadArtifact(context.Background(), owner, repo, artifact.GetID(), 1)
			if err != nil {
				log.Fatal(err)
			}

			f, err := os.CreateTemp("", "*.zip")
			if err != nil {
				log.Fatal(err)
			}
			defer os.Remove(f.Name())

			resp, err := http.Get(url.String())
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()

			if _, err := io.Copy(f, resp.Body); err != nil {
				log.Fatal(err)
			}

			if err := f.Close(); err != nil {
				log.Fatal(err)
			}

			ghRepo := fmt.Sprintf("%s/%s", owner, repo)
			if err := processArtifact(f.Name(), ghRepo); err != nil {
				log.Fatal(err)
			}
		}
	}
}
