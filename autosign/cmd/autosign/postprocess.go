package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func postProcess(config Config) error {
	if config.S3.Endpoint != "" {
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

		files, err := os.ReadDir(config.OutputDir)
		if err != nil {
			return fmt.Errorf("failed reading output directory: %v", err)
		}

		for _, file := range files {
			// skip directories, we only want regular files
			if file.IsDir() {
				continue
			}

			filepath := filepath.Join(config.OutputDir, file.Name())

			info, err := client.FPutObject(ctx, config.S3.Bucket, file.Name(), filepath, minio.PutObjectOptions{
				ContentType: "application/octet-stream",
			})
			if err != nil {
				return fmt.Errorf("error uploading file: %v", err)
			}

			err = os.Remove(filepath)
			if err != nil {
				return fmt.Errorf("error removing file: %v", err)
			}

			log.Printf("Successfully uploaded %s of size %d\n", file.Name(), info.Size)
		}
	}

	err := triggerWebhook(config)
	if err != nil {
		return fmt.Errorf("error triggering webhook: %v", err)
	}

	return nil
}
