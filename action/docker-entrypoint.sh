#!/usr/bin/bash
set -e

export REPONAME=geschenkerbauer
export BUILDDIR=/build
export PKGDEST=$GITHUB_WORKSPACE/repo
export SRCDEST=/srcdest
export SRCPKGDEST=/srcpkgdest
export HOME=/build

# this is required to build some packages
export SHELL=/bin/bash

# create a temporary gnupg keyring
if [[ -n "$GNUPG_PUBKEYRING" ]]; then
    gpg2 --import "$GNUPG_PUBKEYRING"
fi

sudo pacman -Syu --noconfirm

for pkg in $@; do
    cd $GITHUB_WORKSPACE/buildsrc/$pkg
    makepkg_args="-is --noconfirm $@"
    makepkg $makepkg_args
done
