#!/bin/sh

[[ -z "$buildsrcdir" ]] && echo "missing buildsrcdir" && exit 1
[[ -z "$repodir" ]] && echo "missing repodir" && exit 1
[[ -z "$PACKAGER" ]] && echo "missing PACKAGER" && exit 1

buildimg=${buildimg:-mutantmonkey/geschenkerbauer:latest}

for pkg in $@; do
    echo "$pkg"

    if [[ -n "$gpgdir" ]]; then
        docker run --rm \
            -v "$buildsrcdir/$pkg":/buildsrc \
            -v "$gpgdir":/gnupg \
            -v "$repodir":/repo \
            -e "GNUPGHOME=/gnupg" \
            -e PACKAGER \
            -e SOURCE_DATE_EPOCH \
            $buildimg
    elif [[ -n "$gpgkeyring" ]]; then
        docker run --rm \
            -v "$buildsrcdir/$pkg":/buildsrc \
            -v "$repodir":/repo \
            -v "$buildsrcdir/keyring.asc":/keyring.asc:ro \
            -e "GNUPG_PUBKEYRING=/keyring.asc" \
            -e PACKAGER \
            -e SOURCE_DATE_EPOCH \
            $buildimg
    else
        docker run --rm \
            -v "$buildsrcdir/$pkg":/buildsrc \
            -v "$repodir":/repo \
            -e PACKAGER \
            -e SOURCE_DATE_EPOCH \
            $buildimg
    fi

    echo "Docker returned exit code $?"
done
