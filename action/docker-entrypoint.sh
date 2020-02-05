#!/usr/bin/bash
set -e

export REPONAME=geschenkerbauer
export BUILDDIR=/build
export PKGDEST=$GITHUB_WORKSPACE/repo
export SRCDEST=/srcdest
export SRCPKGDEST=/srcpkgdest

# this is required to build some packages
export SHELL=/bin/bash

# create a temporary gnupg keyring
export GNUPGHOME=/build/.gnupg
if [[ -n "$GNUPG_PUBKEYRING" ]]; then
    gpg2 --import "$GNUPG_PUBKEYRING"
fi

sudo pacman -Syu --noconfirm

cd $(mktemp -d /var/tmp/buildsrc-XXXXXXXXXX)
cp -a $GITHUB_WORKSPACE/buildsrc/. .

makepkg_args="-s --noconfirm $@"
makepkg $makepkg_args
