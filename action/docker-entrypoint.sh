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

# unify timestamps so builds can be reproducible
if [[ ! -v SOURCE_DATE_EPOCH ]]; then
	export SOURCE_DATE_EPOCH=$(date +%s)
fi

sudo pacman -Syu --noconfirm

# create associate array that we will use for mapping pkgname to pkgbase
declare -A pkgname_to_pkgbase

function get_deptree_for_pkg {
    srcinfo="${pkgname_to_pkgbase["$1"]}/.SRCINFO"
    [ -f "$srcinfo" ] || return 1

    deps="$(grep -P '^\t(make)?depends' "$srcinfo" | sed 's/^[^\S=]\+ = \([^<>= ]\+\).*$/\1/g')"

    for dep in $deps; do
        # check that dep is a different package to prevent infinite recursion
        if [[ "$dep" != "$1" ]]; then
            get_deptree_for_pkg "$dep"
        fi
    done
    echo "$1"
}

cd "$GITHUB_WORKSPACE/buildsrc"

if [[ "$INPUT_NODEPS" == "1" ]] || [[ "$INPUT_NODEPS" == "true" ]]; then
    echo "::warning::Dependency checking skipped"

    # install git because some packages may need it to even download sources
    sudo pacman -S --noconfirm git

    cp -a "$1/." /startdir

    pushd /startdir
    makepkg --nodeps --noconfirm
    popd
else
    # inspect packages in the current directory and populate the
    # pkgname_to_pkgbase associative array
    for srcinfo in */.SRCINFO; do
        pkgbase="${srcinfo/\/.SRCINFO/}"
        for pkgname in $(grep -P '^pkgname' "$srcinfo" | sed 's/^[^\S=]\+ = \([^<>= ]\+\).*$/\1/g'); do
            pkgname_to_pkgbase["$pkgname"]="$pkgbase"
        done
    done

    for mainpkg in "$@"; do
        for pkg in $(get_deptree_for_pkg "$mainpkg"); do
            pkgbase="${pkgname_to_pkgbase["$pkg"]}"
            echo "::group::${pkgbase}"

            find /startdir -mindepth 1 -delete
            cp -a "${pkgbase}/." /startdir

            pushd /startdir
            makepkg -is --noconfirm
            popd

            echo "::endgroup::"
        done
    done
fi
