#!/usr/bin/bash
export GITHUB_USER="${GITHUB_ACTOR}"
export GITHUB_TOKEN="${INPUT_GITHUB_TOKEN}"

git config user.name "${INPUT_AUTHOR_NAME}"
git config user.email "${INPUT_AUTHOR_EMAIL}"
git remote set-url origin "https://${GITHUB_ACTOR}:${GITHUB_TOKEN}@github.com/${GITHUB_REPOSITORY}"

python update_check_pypi.py
