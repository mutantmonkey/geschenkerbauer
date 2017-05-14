#!/bin/bash

. "$(dirname "$0")"/config.sh

[[ -z "$buildhost" ]] && echo "missing buildhost" && exit 1
[[ -z "$buildsrcdir" ]] && echo "missing buildsrcdir" && exit 1
[[ -z "$repodir" ]] && echo "missing repodir" && exit 1
[[ -z "$gpgdir" ]] && echo "missing gpgdir" && exit 1
[[ -z "$PACKAGER" ]] && echo "missing PACKAGER" && exit 1

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

    ssh $buildhost -t docker run --rm -it -v "$buildsrcdir/$1":/buildsrc -v "$gpgdir":/gnupg -v "$repodir":/repo -e "PACKAGER='$PACKAGER'" quay.io/mutantmonkey/geschenkerbauer:latest
    [ $? -eq 0 ] && builtpkgs[$1]=1
}

function build_all() {
    for dir in *; do
        [ -f "$dir/PKGBUILD" ] && build_deptree "$dir"
    done
}

rsync -avP . $buildhost:$buildsrcdir

if [ -n "$1" ]; then
    [ -d "$1" ] || (echo "Package $1 does not exist in the current directory." && exit 1)
    load_packages
    build_deptree "$1"
else
    load_packages
    build_all
fi
