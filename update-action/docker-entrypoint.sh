#!/usr/bin/bash
export GITHUB_USER="${GITHUB_ACTOR}"
export GITHUB_TOKEN="${INPUT_GITHUB_TOKEN}"

git config user.name "${INPUT_AUTHOR_NAME}"
git config user.email "${INPUT_AUTHOR_EMAIL}"
git remote set-url origin "https://${GITHUB_ACTOR}:${GITHUB_TOKEN}@github.com/${GITHUB_REPOSITORY}"

cd "${GITHUB_WORKSPACE}"
for package in $(ls */PKGBUILD | sed 's/\/PKGBUILD$//g'); do
    export branch_name="aur-updates/${package}/$(date --utc +%s)"

    git checkout -B "${branch_name}"
    git subtree pull -P "${package}" https://aur.archlinux.org/${package}.git master -m "Merge subtree '${package}'"
    if [[ -n "$(git diff --name-only master)" ]]; then
        if [ -n "$(git diff --name-only --diff-filter=U)" ]; then
            echo "::warning Skipping ${package} due to merge conflicts"
            git merge --abort
        else
            git push origin "${branch_name}"
            hub pull-request -m "Update ${package}" --no-edit
        fi
        git checkout master
    fi
done
