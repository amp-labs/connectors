#!/bin/bash

#
# Referenced from Makefile
#
pr_template() {
  SOURCE_BRANCH=$(git rev-parse --abbrev-ref HEAD); # current branch from which PR is created
  TARGET_BRANCH="main"

  USER_NAME=$(git config user.name)
  REPO_NAME=$(basename -s .git `git config --get remote.origin.url`) # your repo name, can be fork name

  echo "PR templates"
  # For every template markdown file construct a URL
  for FILE_NAME in ".github/PULL_REQUEST_TEMPLATE"/*
  do
    TEMPLATE=$(basename "$FILE_NAME")
    # Construct URL for comparing branch against main origin
    PR_URL="https://github.com/amp-labs/connectors/compare/${TARGET_BRANCH}...${USER_NAME}:${REPO_NAME}:${SOURCE_BRANCH}?template=${TEMPLATE}"

    # Display 2 columns with ident, where first column is min 40 chars
    printf "\t %-40s %--s\n" "${TEMPLATE}:" "$PR_URL"
  done
}
