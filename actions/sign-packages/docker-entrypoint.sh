#!/usr/bin/bash
set -e

install -d -m 0700 /tmp/signify
echo "$INPUT_SIGNIFY_SECRET_KEY" > /tmp/signify/pkg.sec

cd $GITHUB_WORKSPACE/repo
sha512sum --tag *.pkg.tar.* > SHA512
signify -S -e -s /tmp/signify/pkg.sec -m SHA512 -x SHA512.sig <<< "$INPUT_SIGNIFY_PASSWORD"
