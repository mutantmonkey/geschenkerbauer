package attestation

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/google/go-github/v66/github"
	"github.com/sigstore/sigstore-go/pkg/bundle"
	"github.com/sigstore/sigstore-go/pkg/root"
	"github.com/sigstore/sigstore-go/pkg/verify"
)

const (
	artifactDigestAlgorithm = "sha256"
	expectedIssuer          = "https://token.actions.githubusercontent.com"
)

//go:embed trusted_root.json
var trustedRootJSON []byte

func VerifyAttestation(filename string, client *github.Client, owner string, repo string) (bool, error) {
	f, err := os.Open(filename)
	if err != nil {
		return false, fmt.Errorf("failed to open file for reading: %v", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return false, fmt.Errorf("failed to read file: %v", err)
	}
	digest := h.Sum(nil)

	attestations, _, err := client.Repositories.ListAttestations(context.Background(), owner, repo, fmt.Sprintf("%s:%x", artifactDigestAlgorithm, digest), nil)
	if err != nil {
		return false, fmt.Errorf("failed to fetch attestations: %v", err)
	}

	if len(attestations.Attestations) == 0 {
		return false, fmt.Errorf("no attestations found")
	}

	// set these up so they should essentially match what gh does
	sev, err := getSignedEntityVerifier()
	if err != nil {
		return false, fmt.Errorf("failed to get SEV: %v", err)
	}

	sanRegex := fmt.Sprintf("^https://github.com/%s/%s/", url.PathEscape(owner), url.PathEscape(repo))
	pb, err := getPolicyBuilder(sanRegex, digest)
	if err != nil {
		return false, fmt.Errorf("failed to get policy builder: %v", err)
	}

	var b *bundle.Bundle
	for _, attestation := range attestations.Attestations {
		if err := json.Unmarshal(attestation.Bundle, &b); err != nil {
			return false, fmt.Errorf("failed to parse attestation: %v", err)
		}

		if _, err := sev.Verify(b, *pb); err != nil {
			return false, fmt.Errorf("failed to verify attestation: %v", err)
		} else {
			return true, nil
		}
	}

	return false, nil
}

func getTrustedMaterial() (root.TrustedMaterialCollection, error) {
	trustedRoot, err := root.NewTrustedRootFromJSON(trustedRootJSON)
	if err != nil {
		return nil, err
	}

	trustedMaterial := root.TrustedMaterialCollection{
		trustedRoot,
	}

	return trustedMaterial, nil
}

func getSignedEntityVerifier() (*verify.SignedEntityVerifier, error) {
	verifierConfig := []verify.VerifierOption{
		verify.WithSignedCertificateTimestamps(1),
		verify.WithObserverTimestamps(1),
		verify.WithTransparencyLog(1),
	}

	trustedMaterial, err := getTrustedMaterial()
	if err != nil {
		return nil, err
	}

	return verify.NewSignedEntityVerifier(trustedMaterial, verifierConfig...)
}

func getIdentityPolicies(sanRegex string) ([]verify.PolicyOption, error) {
	certID, err := verify.NewShortCertificateIdentity(expectedIssuer, "", "", sanRegex)
	if err != nil {
		return nil, err
	}

	return []verify.PolicyOption{
		verify.WithCertificateIdentity(certID),
	}, nil
}

func getPolicyBuilder(sanRegex string, digest []byte) (*verify.PolicyBuilder, error) {
	identityPolicies, err := getIdentityPolicies(sanRegex)
	if err != nil {
		return nil, err
	}

	artifactPolicy := verify.WithArtifactDigest(artifactDigestAlgorithm, digest)

	pb := verify.NewPolicy(artifactPolicy, identityPolicies...)
	return &pb, nil
}
