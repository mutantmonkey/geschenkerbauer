#!/bin/sh
#image=docker.io/archlinux/archlinux:base-devel
image=ghcr.io/archlinux/archlinux:base-devel
digest=$(cosign verify $image --certificate-identity-regexp="https://gitlab\.archlinux\.org/archlinux/archlinux-docker//\.gitlab-ci\.yml@refs/tags/v[0-9]+\.0\.[0-9]+" --certificate-oidc-issuer=https://gitlab.archlinux.org | jq -r ".[0].critical.image[\"docker-manifest-digest\"]")
sed -i "s#^FROM .*\$#FROM ${image}@${digest}#" action/Dockerfile
