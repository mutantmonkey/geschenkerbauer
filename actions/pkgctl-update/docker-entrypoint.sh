#!/bin/bash
set -euo pipefail

export TERM=${TERM:-xterm-256color}

mkdir -p /startdir
cp -a "$1/." /startdir
chown -R builduser:users /startdir

pushd /startdir

sudo -u builduser pkgctl version upgrade .
sudo -u builduser makepkg --printsrcinfo > .SRCINFO

popd

cp /startdir/PKGBUILD "$1/PKGBUILD"
cp /startdir/.SRCINFO "$1/.SRCINFO"
