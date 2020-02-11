#!/bin/bash
# This script unpacks packages from a ZIP file, optionally signs them, and adds
# them to a repository. This is useful for working with the artifact files
# generated by GitHub Actions.

OUTPUT_TMPDIR=$(mktemp -d)
. repo_config.sh

unzip -d $OUTPUT_TMPDIR $1
pushd $OUTPUT_TMPDIR >/dev/null

for f in *.pkg.tar.*; do
    # GitHub Actions forbids : in filenames, so the build action replaces them
    # before creating the ZIP. Now that we have the file, rename it back.
    if [[ "$f" == *"__3A__"* ]]; then
        old_filename="$f"
        f="${f/__3A__/:}"
        mv "$old_filename" "$f"
    fi

    if [ ! -f "$OUTPUT_REPO/$f" ]; then
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

popd >/dev/null
rmdir "$OUTPUT_TMPDIR"
