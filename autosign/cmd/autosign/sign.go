package main

import (
	"os"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
)

func signPackage(filename string, keyring string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	sigf, err := os.Create(filename + ".sig")
	if err != nil {
		return err
	}
	defer sigf.Close()

	keyringFile, err := os.Open(keyring)
	if err != nil {
		return err
	}
	defer keyringFile.Close()

	reader := packet.NewReader(keyringFile)
	entity, err := openpgp.ReadEntity(reader)
	if err != nil {
		return err
	}

	if err := openpgp.DetachSign(sigf, entity, f, nil); err != nil {
		return err
	}

	return nil
}
