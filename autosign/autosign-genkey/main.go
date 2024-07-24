package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

const (
	keyType = "rsa"
	keyBits = 4096
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Provide the destination path as the sole argument; e.g. autosign-genkey output.pgp\n")
		os.Exit(1)
	}
	destPath := os.Args[1]

	reader := bufio.NewReader(os.Stdin)
	var err error

	name := os.Getenv("SIGNER_NAME")
	if name == "" {
		fmt.Print("Signer's name: ")
		name, err = reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
	}
	name = strings.TrimSpace(name)

	email := os.Getenv("SIGNER_EMAIL")
	if email == "" {
		fmt.Print("Email address: ")
		email, err = reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
	}
	email = strings.TrimSpace(email)

	if name == "" {
		fmt.Fprintf(os.Stderr, "A name must be provided either when prompted or with the SIGNER_NAME environment variable\n")
		os.Exit(2)
	}
	if email == "" {
		fmt.Fprintf(os.Stderr, "An email address must be provided either when prompted or with the SIGNER_NAME environment variable\n")
		os.Exit(2)
	}

	fmt.Fprintf(os.Stderr, "Signer identity: %s <%s>\n", name, email)

	newKey, err := crypto.GenerateKey(name, email, keyType, keyBits)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create(destPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	output, err := newKey.Serialize()
	if err != nil {
		log.Fatal(err)
	}

	f.Write(output)
}
