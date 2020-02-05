#!/usr/bin/bash
set -e

export REPONAME=geschenkerbauer
export BUILDDIR=/build
export PKGDEST=/repo
export SRCDEST=/srcdest
export SRCPKGDEST=/srcpkgdest
export INPUT_REPODIR=${INPUT_REPODIR:-/repo}
export INPUT_BUILDSRC=${INPUT_BUILDSRC:-/buildsrc}

# this is required to build some packages
export SHELL=/bin/bash

# create a temporary gnupg keyring if an existing one was not provided
if [[ -z "$GNUPGHOME" ]]; then
    export GNUPGHOME=/build/.gnupg
    if [[ -n "$GNUPG_PUBKEYRING" ]]; then
        gpg2 --import "$GNUPG_PUBKEYRING"
    fi
fi

# create an empty repo database if one does not exist
[ -e $INPUT_REPODIR/$REPONAME.db ] || touch $INPUT_REPODIR/$REPONAME.db

sudo pacman -Syu --noconfirm

cd $(mktemp -d /var/tmp/buildsrc-XXXXXXXXXX)
cp -a $INPUT_BUILDSRC/. .

makepkg_args="-s --noconfirm $@"
makepkg $makepkg_args

if [ $? -eq 0 ]; then
    for f in $(makepkg --packagelist); do
        repo-add $INPUT_REPODIR/$REPONAME.db.tar.gz $f
    done
fi
