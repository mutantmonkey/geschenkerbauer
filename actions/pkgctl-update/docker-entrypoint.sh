#!/bin/sh
set -e
pushd $1
pkgctl version upgrade .
sudo -u builduser BUILDDIR=/tmp PKGDEST=/tmp SRCDEST=/tmp makepkg --printsrcinfo > .SRCINFO
popd
