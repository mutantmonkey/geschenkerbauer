#!/bin/bash

[[ -z "$buildsrcdir" ]] && echo "missing buildsrcdir" && exit 1
[[ -z "$repodir" ]] && echo "missing repodir" && exit 1
[[ -z "$gpgdir" ]] && echo "missing gpgdir" && exit 1
[[ -z "$PACKAGER" ]] && echo "missing PACKAGER" && exit 1

for pkg in $(</dev/stdin); do
    docker run --rm -v "$buildsrcdir/$pkg":/buildsrc -v "$gpgdir":/gnupg -v "$repodir":/repo -e "GNUPGHOME=/gnupg" -e "PACKAGER='$PACKAGER'" quay.io/mutantmonkey/geschenkerbauer:latest
    echo "Docker returned exit code $?"
    #[ $? -eq 0 ] || exit 1
done
