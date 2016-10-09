#!/bin/bash

buildhost=coreos1
buildsrcdir=/tmp/geschenkerbauer.cBCw5WkbOy
repodir=/home/core/arch/repo
gpgdir=/home/core/arch/gnupg
PACKAGER="mutantmonkey <archpkg@mutantmonkey.mx>"

declare -A pkgbases
declare -A builtpkgs

function load_packages() {
    echo "Loading packages..."

    for dir in *; do
        [ -f "$dir/PKGBUILD" ] || continue

        pushd "$dir" >/dev/null
        pkgs=($(grep pkgname .SRCINFO | sed 's/^\w\+ = \(.*\)$/\1/g'))
        popd >/dev/null

        for pkg in "${pkgs[@]}"; do
            pkgbases[$pkg]="$dir"
        done
    done
}

function build_deptree() {
    echo "Starting build of $1..."

    pushd "$1" >/dev/null
    pkgs=($(grep $'\t''\(checkdepends\|depends\|makedepends\)' .SRCINFO | sed 's/^\t\w\+ = \([^<>=]\+\).*$/\1/' | uniq))
    popd >/dev/null

    for pkg in "${pkgs[@]}"; do
        pkgbase="${pkgbases[$pkg]}"
        [ -f "$pkgbase/PKGBUILD" ] && [ -z "${builtpkgs[$pkgbase]}" ] && \
            [ "$pkgbase" != "$1" ] && \
            build_deptree "$pkgbase"
    done

    ssh $buildhost -t docker run --rm -it -v "$buildsrcdir/$1":/buildsrc -v "$gpgdir":/gnupg -v "$repodir":/repo -e "PACKAGER='$PACKAGER'" geschenkerbauer
    [ $? -eq 0 ] && builtpkgs[$1]=1
}

function build_all() {
    for dir in *; do
        [ -f "$dir/PKGBUILD" ] && build_deptree "$dir"
    done
}

rsync -avP ~/arch/packages/ $buildhost:$buildsrcdir

if [ -n "$1" ]; then
    [ -d "$1" ] || (echo "Package $1 does not exist in the current directory." && exit 1)
    load_packages
    build_deptree "$1"
else
    load_packages
    build_all
fi
