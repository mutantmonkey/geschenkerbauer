#!/bin/sh

[[ -z "$buildsrcdir" ]] && echo "missing buildsrcdir" && exit 1
[[ -z "$repodir" ]] && echo "missing repodir" && exit 1
[[ -z "$gpgdir" ]] && echo "missing gpgdir" && exit 1
[[ -z "$PACKAGER" ]] && echo "missing PACKAGER" && exit 1

for pkg in $@; do
    echo "$pkg"
    docker run --rm \
        -v "$buildsrcdir/$pkg":/buildsrc \
        -v "$gpgdir":/gnupg \
        -v "$repodir":/repo \
        -e "GNUPGHOME=/gnupg" \
        -e "PACKAGER='$PACKAGER'" \
        mutantmonkey/geschenkerbauer:latest
    echo "Docker returned exit code $?"
    #[ $? -eq 0 ] || exit 1
done
