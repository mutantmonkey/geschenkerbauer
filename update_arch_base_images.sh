#!/bin/bash
set -eu

function get_attested_image() {
    local image=$1
    local digest=$(cosign verify $1 --certificate-identity-regexp="https://gitlab\.archlinux\.org/archlinux/archlinux-docker//\.gitlab-ci\.yml@refs/tags/v[0-9]+\.0\.[0-9]+" --certificate-oidc-issuer=https://gitlab.archlinux.org | jq -r ".[0].critical.image[\"docker-manifest-digest\"]")
    echo "${image}@${digest}"
}


arch_image_repo=ghcr.io/archlinux/archlinux
base_image=$(get_attested_image ${arch_image_repo}:base)
devel_image=$(get_attested_image ${arch_image_repo}:base-devel)

sed -i "s#^FROM ${arch_image_repo}.*\$#FROM ${base_image}#" containers/autosign-receiver/Dockerfile
sed -i "s#^FROM ${arch_image_repo}.*\$#FROM ${devel_image}#" containers/builder/Dockerfile

git add containers/{autosign-receiver,builder}/Dockerfile
git commit -m "Update Arch Linux Docker base images"
