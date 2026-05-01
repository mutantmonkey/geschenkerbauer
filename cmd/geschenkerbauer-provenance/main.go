package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v6"

	"mutantmonkey.in/code/geschenkerbauer/pkg/provenance"
)

func main() {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// read package names from environment
	pkgs := strings.Fields(os.Getenv("GESCHENKERBAUER_PACKAGES"))

	r, err := git.PlainOpen(path)
	if err != nil {
		log.Fatal(err)
	}

	resourceDescriptor, err := provenance.GetResourceDescriptor(r)
	if err != nil {
		log.Fatal(err)
	}

	p := &provenance.Provenance{
		BuildDefinition: &provenance.BuildDefinition{
			BuildType:          provenance.BuilderUri,
			ExternalParameters: &provenance.ExternalParameters{},
			InternalParameters: &provenance.InternalParameters{
				BuilderParams: &provenance.BuilderParams{
					Packages: pkgs,
				},
			},
			ResolvedDependencies: []*provenance.ResourceDescriptor{
				resourceDescriptor,
			},
		},
		RunDetails: &provenance.RunDetails{
			Builder: &provenance.Builder{
				Id: provenance.BuilderUri,
			},
		},
	}

	statement := &provenance.Statement{
		Type:          provenance.StatementTypeUri,
		Subject:       []*provenance.Subject{},
		PredicateType: provenance.PredicateTypeUri,
		Predicate:     p,
	}

	for _, subjectPath := range os.Args[1:] {
		h := sha256.New()

		f, err := os.Open(subjectPath)
		if err != nil {
			log.Fatal(err)
		}

		if _, err := io.Copy(h, f); err != nil {
			log.Fatal(err)
		}

		f.Close()

		subject := &provenance.Subject{
			Name: filepath.Base(subjectPath),
			Digest: map[string]string{
				"sha256": fmt.Sprintf("%x", h.Sum(nil)),
			},
		}

		statement.Subject = append(statement.Subject, subject)
	}

	if len(statement.Subject) <= 0 {
		log.Fatal("No subject(s) were specified as an argument")
	}

	data, err := json.Marshal(statement)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", data)

	// After we have the DSSE envelope, it must be signed
	// we can provide a sigstore bundle or whatever
}
