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
    elif [[ -n "$gpgkeyring" ]]; then
        docker run --rm \
            -v "$buildsrcdir/$pkg":/buildsrc \
            -v "$repodir":/repo \
            -v "$buildsrcdir/keyring.asc":/keyring.asc:ro \
            -e "GNUPG_PUBKEYRING=/keyring.asc" \
            -e PACKAGER \
            mutantmonkey/geschenkerbauer:latest
    else
        docker run --rm \
            -v "$buildsrcdir/$pkg":/buildsrc \
            -v "$repodir":/repo \
            -e PACKAGER \
            mutantmonkey/geschenkerbauer:latest
    fi

    echo "Docker returned exit code $?"
done
