package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"mutantmonkey.in/code/geschenkerbauer/autosign/internal/fshelpers"
)

func downloadFromS3(config Config) error {
	ctx := context.Background()
	client, err := minio.New(config.S3.Endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			config.S3.AccessKeyID,
			config.S3.SecretAccessKey,
			""),
		Secure: true,
	})
	if err != nil {
		return fmt.Errorf("error creating S3 client: %v", err)
	}

	for object := range client.ListObjects(ctx, config.S3.Bucket, minio.ListObjectsOptions{}) {
		if object.Err != nil {
			return fmt.Errorf("error listing objects: %v", err)
		}

		filename := filepath.Base(filepath.Clean(object.Key))
		destpath := filepath.Join(config.IncomingDir, filename)

		if _, err := os.Stat(filepath.Join(config.RepoDir, filename)); err == nil {
			log.Printf("Warning: skipping %q because it already exists in the output directory", filename)
			continue
		}

		if err := client.FGetObject(ctx, config.S3.Bucket, filename, destpath, minio.GetObjectOptions{}); err != nil {
			return fmt.Errorf("failed to download object: %v", err)
		}
	}

	return nil
}

func ProcessIncoming(config Config) error {
	if config.S3.Endpoint != "" {
		if err := downloadFromS3(config); err != nil {
			return fmt.Errorf("failed downloading from S3: %v", err)
		}
	}

	repo := fmt.Sprintf("%s/%s", config.GitHub.Owner, config.GitHub.Repo)
	fileSystem := os.DirFS(config.IncomingDir)

	files, err := fs.Glob(fileSystem, "*.pkg.tar.zst")
	if err != nil {
		return fmt.Errorf("failed reading incoming directory: %v", err)
	}

	for _, filename := range files {
		filename = filepath.Base(filepath.Clean(filename))
		incomingFilepath := filepath.Join(config.IncomingDir, filename)
		fmt.Printf("%s\n", incomingFilepath)

		// skip packages that already exist in the output directory
		if _, err := os.Stat(filepath.Join(config.RepoDir, filename)); err == nil {
			log.Printf("Warning: skipping %q because it already exists in the output directory", filename)
			continue
		}

		if _, err := os.Stat(incomingFilepath + ".sig"); err != nil {
			log.Printf("Warning: skipping %q because signature was not present", filename)
			continue
		}

		// verify attestation
		cmd := exec.Command("gh", "attestation", "verify", incomingFilepath, "-R", repo)
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, fmt.Sprintf("GH_TOKEN=%s", config.GitHub.AuthToken))
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error validating attestation: %v", err)
		}

		// move package signature to final output directory
		if err := fshelpers.MoveFile(incomingFilepath+".sig", filepath.Join(config.RepoDir, filename+".sig")); err != nil {
			return fmt.Errorf("error moving package signature to destination directory: %v", err)
		}

		// move package to final output directory
		if err := fshelpers.MoveFile(incomingFilepath, filepath.Join(config.RepoDir, filename)); err != nil {
			return fmt.Errorf("error moving package to destination directory: %v", err)
		}

		// add new packages to repository database
		cmd = exec.Command("repo-add", config.DbName, filename)
		cmd.Dir = config.RepoDir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error adding package to repository database: %v", err)
		}
	}

	return nil
}
