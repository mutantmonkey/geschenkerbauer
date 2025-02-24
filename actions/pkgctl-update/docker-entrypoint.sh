#!/bin/sh
set -e
pkgctl version upgrade .
makepkg --printsrcinfo > .SRCINFO
