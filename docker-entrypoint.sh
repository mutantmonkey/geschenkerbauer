#!/usr/bin/bash
set -e

export REPONAME=geschenkerbauer
export BUILDDIR=/build
export PKGDEST=$(mktemp -d /var/tmp/pkgdest-XXXXXXXXXX)
export SRCDEST=/srcdest
export SRCPKGDEST=/srcpkgdest

# this is required to build some packages
export SHELL=/bin/bash

# create a temporary gnupg keyring if an existing one was not provided
if [[ -z "$GNUPGHOME" ]]; then
    export GNUPGHOME=/gnupg
    if [[ -n "$GNUPG_PUBKEYRING" ]]; then
        gpg2 --import "$GNUPG_PUBKEYRING"
    fi
fi

# create an empty repo database if one does not exist
[ -e /repo/$REPONAME.db ] || touch /repo/$REPONAME.db

sudo pacman -Syu --noconfirm

cd $(mktemp -d /var/tmp/buildsrc-XXXXXXXXXX)
cp -a /buildsrc/. .

makepkg_args="-s --noconfirm $@"
makepkg $makepkg_args

if [ $? -eq 0 ]; then
    cd "$PKGDEST"
    for f in *; do
        cp $f /repo/
        repo-add /repo/$REPONAME.db.tar.gz /repo/$f
    done
fi
