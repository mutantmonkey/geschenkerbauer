#!/usr/bin/bash
set -e

export REPONAME=geschenkerbauer
export BUILDDIR=/build
export PKGDEST=/repo
export SRCDEST=/srcdest
export SRCPKGDEST=/srcpkgdest

# this is required to build some packages
export SHELL=/bin/bash

# create an empty repo database if one does not exist
[ -e /repo/$REPONAME.db ] || touch /repo/$REPONAME.db

sudo pacman -Syu --noconfirm

cd $(mktemp -d /var/tmp/buildsrc-XXXXXXXXXX)
cp -a /buildsrc/. .

makepkg_args="-s --noconfirm $@"
makepkg $makepkg_args

if [ $? -eq 0 ]; then
    for f in *.pkg.tar.xz; do
        repo-add /repo/$REPONAME.db.tar.gz /repo/$f
    done
fi
