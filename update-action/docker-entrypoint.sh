#!/usr/bin/bash
export GITHUB_USER="${GITHUB_ACTOR}"
export GITHUB_TOKEN="${INPUT_GITHUB_TOKEN}"

git config user.name "${INPUT_AUTHOR_NAME}"
git config user.email "${INPUT_AUTHOR_EMAIL}"
git remote set-url origin "https://${GITHUB_ACTOR}:${GITHUB_TOKEN}@github.com/${GITHUB_REPOSITORY}"

for package in $(ls */PKGBUILD | sed 's/\/PKGBUILD$//g'); do
    export branch_name="aur-updates/${package}/$(date --utc +%s)"

    git checkout -B "${branch_name}"
    git subtree pull -P "${package}" https://aur.archlinux.org/${package}.git master -m "Merge subtree '${package}'"
    if [[ "$(git rev-parse "${branch_name}")" != "$(git rev-parse master)" ]]; then
        git push origin "${branch_name}"
        hub pull-request -m "Update ${package}" --no-edit
        git checkout master
    fi
done
