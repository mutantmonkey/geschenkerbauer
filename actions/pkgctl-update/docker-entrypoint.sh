#!/bin/bash
set -euo pipefail

cp -a "$1/." /startdir
chown -R builduser:users /startdir

pushd /startdir

sudo -u builduser pkgctl version upgrade .
sudo -u builduser BUILDDIR=/tmp PKGDEST=/tmp SRCDEST=/tmp makepkg --printsrcinfo > .SRCINFO

popd

cp /startdir/PKGBUILD "$1/PKGBUILD"
cp /startdir/.SRCINFO "$1/.SRCINFO"
