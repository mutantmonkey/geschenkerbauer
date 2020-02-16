#!/usr/bin/bash
export GITHUB_USER="${GITHUB_ACTOR}"
export GITHUB_TOKEN="${INPUT_GITHUB_TOKEN}"

git config user.name "${INPUT_AUTHOR_NAME}"
git config user.email "${INPUT_AUTHOR_EMAIL}"
git remote set-url origin "https://${GITHUB_ACTOR}:${GITHUB_TOKEN}@github.com/${GITHUB_REPOSITORY}"

cd "${GITHUB_WORKSPACE}" || exit 1

# get a list of existing pull requests and map them to package names
declare -A branch_by_package
while IFS=$'\n' read -ra line; do
    read -ra entry <<< "${line/,Update/ }"
    branch_by_package+=([${entry[0]}]=${entry[2]})
done <<< "$(hub pr list -f '%I,%t %H%n')"

for package in $(find -- */PKGBUILD | sed 's/\/PKGBUILD$//g'); do
    branch_name="aur-updates/${package}/$(date --utc +%s)"

    git checkout -B "${branch_name}"
    git subtree pull -P "${package}" "https://aur.archlinux.org/${package}.git" master -m "Merge subtree '${package}'"
    if [[ -n "$(git diff --name-only master)" ]]; then
        if [ -n "$(git diff --name-only --diff-filter=U)" ]; then
            echo "::warning::Skipping ${package} due to merge conflicts"
            git merge --abort
        else
            # check if there is an existing branch that we can use
            # if not, push the current branch and create a new pull request
            # if there is, force push to it so the PR will be updated
            existing_branch="${branch_by_package[${package}]}"
            if [[ -z "${existing_branch}" ]]; then
                git push origin "${branch_name}"
                hub pull-request -m "Update ${package}" --no-edit
            elif [[ -n "$(git diff --name-only "${existing_branch}")" ]]; then
                git push -f origin "${existing_branch}"
            fi
        fi
        git checkout master
    fi
done
