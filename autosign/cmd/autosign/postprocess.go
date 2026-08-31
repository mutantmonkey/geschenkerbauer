package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/gokrazy/rsync"
	"github.com/gokrazy/rsync/rsyncclient"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func postProcess(config Config) error {
	ctx := context.Background()

	if config.S3.Endpoint != "" {
		client, err := minio.New(config.S3.Endpoint, &minio.Options{
			Creds: credentials.NewStaticV4(
				config.S3.AccessKeyID,
				config.S3.SecretAccessKey,
				""),
			Secure: true,
		})
		if err != nil {
			return fmt.Errorf("creating S3 client: %v", err)
		}

		files, err := os.ReadDir(config.OutputDir)
		if err != nil {
			return fmt.Errorf("reading output directory: %v", err)
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
				return fmt.Errorf("uploading file: %v", err)
			}

			err = os.Remove(filepath)
			if err != nil {
				return fmt.Errorf("removing file: %v", err)
			}

			log.Printf("Successfully uploaded %s of size %d\n", file.Name(), info.Size)
		}
	} else if config.Rsync.RemotePath != "" {
		keyString, err := os.ReadFile(config.Rsync.SSHKeyPath)
		if err != nil {
			return fmt.Errorf("reading SSH private key: %v", err)
		}

		signer, err := ssh.ParsePrivateKey(keyString)
		if err != nil {
			return fmt.Errorf("parsing SSH private key: %v", err)
		}

		hostKeyCallback, err := knownhosts.New(config.Rsync.KnownHostsPath)
		if err != nil {
			return fmt.Errorf("loading SSH known hosts: %v", err)
		}

		conn, err := ssh.Dial("tcp", config.Rsync.Addr, &ssh.ClientConfig{
			User:            config.Rsync.User,
			Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
			HostKeyCallback: hostKeyCallback,
		})
		if err != nil {
			return fmt.Errorf("dialing SSH: %v", err)
		}

		session, err := conn.NewSession()
		if err != nil {
			return fmt.Errorf("opening SSH session: %v", err)
		}
		defer session.Close()

		stdin, err := session.StdinPipe()
		if err != nil {
			return err
		}

		stdout, err := session.StdoutPipe()
		if err != nil {
			return err
		}

		// The remote authorized keys file will force the command
		// `gokr-rsync --daemon`, so the command here is ignored
		if err := session.Shell(); err != nil {
			return fmt.Errorf("starting remote gokr-rsync: %v", err)
		}

		rw := &rsync.BothCloser{
			ReadCloser:  io.NopCloser(stdout),
			WriteCloser: stdin,
		}

		client, err := rsyncclient.New([]string{"-rlpt"}, rsyncclient.WithSender())
		if err != nil {
			return fmt.Errorf("creating rsync client: %v", err)
		}

		if _, err := client.RunDaemon(ctx, rw, config.Rsync.RemotePath, []string{config.OutputDir + "/"}); err != nil {
			return fmt.Errorf("rsync transfer failed: %v", err)
		}

		if err := session.Wait(); err != nil {
			return fmt.Errorf("SSH error: %v", err)
		}

		// We now have to clean up the temporary files ourselves since
		// gokr-rsync does not implement --remove-source-files

		log.Print("Files copied via rsync, beginning cleanup...")

		files, err := os.ReadDir(config.OutputDir)
		if err != nil {
			return fmt.Errorf("reading output directory: %v", err)
		}

		for _, file := range files {
			// skip directories, we only want regular files
			if file.IsDir() {
				continue
			}

			if err := os.Remove(filepath.Join(config.OutputDir, file.Name())); err != nil {
				return fmt.Errorf("removing file: %v", err)
			}
		}
	}

	err := triggerWebhook(config)
	if err != nil {
		return fmt.Errorf("error triggering webhook: %v", err)
	}

	return nil
}
