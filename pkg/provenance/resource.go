package provenance

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v6"
)

const (
	StatementTypeUri = "https://in-toto.io/Statement/v1"
	PredicateTypeUri = "https://slsa.dev/provenance/v1"
	BuilderUri       = "https://geschenkerbauer.mutantmonkey.in/build/v1"
)

func GetResourceDescriptor(r *git.Repository) (resource *ResourceDescriptor, err error) {
	ref, err := r.Head()
	if err != nil {
		return nil, err
	}

	remote, err := r.Remote("origin")
	if err != nil {
		return nil, err
	}

	// https://git-scm.com/book/en/v2/Git-on-the-Server-The-Protocols

	// This should return something like git+https://example.com/git/example@refs/heads/main
	// https://spdx.github.io/spdx-spec/v2.3/package-information/#77-package-download-location-field

	remoteConfig := remote.Config()
	gitRemoteUrl := remoteConfig.URLs[0]

	// FIXME: is this a safe way to handle user input?
	var remoteUrl string
	if strings.HasPrefix(gitRemoteUrl, "https://") {
		remoteUrl = "git+" + remoteUrl
	} else if strings.HasPrefix(gitRemoteUrl, "ssh://") {
		remoteUrl = "git+" + remoteUrl
	} else if strings.Contains(gitRemoteUrl, ":") {
		gitHost, gitPath, found := strings.Cut(gitRemoteUrl, ":")
		if !found {
			return nil, errors.New("Unsupported Git remote URL format (detected SSH)")
		}

		_, gitHost, _ = strings.Cut(gitHost, "@")
		remoteUrl = fmt.Sprintf("git+ssh://%s/%s", gitHost, gitPath)
	} else {
		return nil, errors.New("Unsupported Git remote URL format")
	}

	resource = &ResourceDescriptor{
		Uri: fmt.Sprintf("%s@%s", remoteUrl, ref.Name()),
		Digest: map[string]string{
			"gitCommit": ref.Hash().String(),
		},
	}

	return
}
