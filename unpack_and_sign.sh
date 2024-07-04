#!/bin/bash
# This script unpacks packages from a ZIP file, optionally signs them, and adds
# them to a repository. This is useful for working with the artifact files
# generated by GitHub Actions.

OUTPUT_TMPDIR=$(mktemp -d)
. repo_config.sh

unzip -d $OUTPUT_TMPDIR $1
pushd $OUTPUT_TMPDIR >/dev/null

# Warn if no source directory is provided
if [ -z "$SOURCE_PATH" ]; then
    echo "No source directory provided, PKGBUILD integrity checking will be skipped."
fi

for f in *.pkg.tar.*; do
    # Verify attestation
    gh attestation verify "$f" -R mutantmonkey/aur
    if [ $? -ne 0 ]; then
        break
    fi

    # GitHub Actions forbids : in filenames, so the build action replaces them
    # before creating the ZIP. Now that we have the file, rename it back.
    if [[ "$f" == *"__3A__"* ]]; then
        old_filename="$f"
        f="${f/__3A__/:}"
        mv "$old_filename" "$f"
    fi

    if [ ! -f "$OUTPUT_REPO/$f" ]; then
        # Verify that pkgbuild_sha256sum in package BUILDINFO matches the
        # sha256sum of the current PKGBUILD in our local source directory
        if [ -n "$SOURCE_PATH" ]; then
            pkgbase=$(tar --force-local -xOf "$f" .BUILDINFO | grep '^pkgbase' | sed 's/^.* = //g')
            expected_sha256sum=$(tar --force-local -xOf "$f" .BUILDINFO | grep '^pkgbuild_sha256sum' | sed 's/^.* = //g')
            actual_sha256sum=$(sha256sum "${SOURCE_PATH}/${pkgbase}/PKGBUILD" | cut -f1 -d ' ')

            # If the sha256sum doesn't match the first time, try adjusting pkgver to match
            if [[ "${expected_sha256sum}" != "${actual_sha256sum}" ]]; then
                pkgver=$(tar --force-local -xOf "$f" .BUILDINFO | grep '^pkgver' | sed 's/^.* = \([^\-]\+\).*$/\1/g')
                actual_sha256sum=$(sed "s/^pkgver=.*$/pkgver=${pkgver}/" "${SOURCE_PATH}/${pkgbase}/PKGBUILD" | sha256sum - | cut -f1 -d ' ')
            fi

            # If the sha256sum still doesn't match, then we're really in trouble
            if [[ "${expected_sha256sum}" != "${actual_sha256sum}" ]]; then
                echo "$f: PKGBUILD sha256sum mismatch!"
                echo "PKGBUILD used for build: ${expected_sha256sum}"
                echo "PKGBUILD in source dir:  ${actual_sha256sum}"
                break
            fi
        fi

        mv "$f" "$OUTPUT_REPO/$f"
        pushd "$OUTPUT_REPO" >/dev/null
        [ -n "$GPG_KEYID" ] && gpg2 --local-user $GPG_KEYID -b "$f"
        repo-add "$OUTPUT_REPO_FILENAME" "$f"
        popd >/dev/null
    else
        rm "$f"
        echo "$f: already exists"
    fi
done

# Delete checksum/signed checksum files if they exist
rm -f SHA512{,.sig}

popd >/dev/null
rmdir "$OUTPUT_TMPDIR"
