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

function get_deptree_for_pkg {
    [ -f "$1/PKGBUILD" ] || return 1

    pushd "$1" >/dev/null
    deps="$(makepkg --printsrcinfo | grep -P '^\t(make)?depends' | sed 's/^[^\S=]\+ = \([^<>= ]\+\).*$/\1/g')"
    popd >/dev/null

    for dep in $deps; do
        get_deptree_for_pkg "$dep"
    done
    echo "$1"
}

cd $GITHUB_WORKSPACE/buildsrc
for mainpkg in $@; do
    for pkg in $(get_deptree_for_pkg $mainpkg); do
        cd $pkg
        makepkg_args="-is --noconfirm $@"
        makepkg $makepkg_args
    done
done
