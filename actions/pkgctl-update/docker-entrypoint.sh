#!/bin/sh
set -e
pushd $1
pkgctl version upgrade .
sudo -u builduser makepkg --printsrcinfo > .SRCINFO
popd
