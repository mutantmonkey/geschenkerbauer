#!/bin/sh

[[ -z "$buildsrcdir" ]] && echo "missing buildsrcdir" && exit 1
[[ -z "$repodir" ]] && echo "missing repodir" && exit 1
[[ -z "$PACKAGER" ]] && echo "missing PACKAGER" && exit 1

for pkg in $@; do
    echo "$pkg"

    if [[ -n "$gpgdir" ]]; then
        docker run --rm \
            -v "$buildsrcdir/$pkg":/buildsrc \
            -v "$gpgdir":/gnupg \
            -v "$repodir":/repo \
            -e "GNUPGHOME=/gnupg" \
            -e "PACKAGER='$PACKAGER'" \
            mutantmonkey/geschenkerbauer:latest
    else
        docker run --rm \
            -v "$buildsrcdir/$pkg":/buildsrc \
            -v "$repodir":/repo \
            -e "PACKAGER='$PACKAGER'" \
            mutantmonkey/geschenkerbauer:latest
    fi

    echo "Docker returned exit code $?"
done
