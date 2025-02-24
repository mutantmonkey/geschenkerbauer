#!/bin/sh
set -e
pushd $1
pkgctl version upgrade .
makepkg --printsrcinfo > .SRCINFO
popd
