#!/usr/bin/bash
set -e

export REPONAME=geschenkerbauer
export BUILDDIR=/build
export PKGDEST="$GITHUB_WORKSPACE/repo"
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

cd "$GITHUB_WORKSPACE/buildsrc"

if [[ "$INPUT_NODEPS" == "1" ]] || [[ "$INPUT_NODEPS" == "true" ]]; then
    echo "::warning::Dependency checking skipped"

    # install git because some packages may need it to even download sources
    sudo pacman -S --noconfirm git

    cd "$1"
    makepkg --noconfirm
else
    for mainpkg in "$@"; do
        for pkg in $(get_deptree_for_pkg "$mainpkg"); do
            pushd "$pkg"
            makepkg -is --noconfirm
            popd
        done
    done
fi
