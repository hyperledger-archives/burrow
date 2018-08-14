#!/usr/bin/env bash
# Gives us a non-zero exit code if there are tracked or untracked changes in the working
# directory
REPO=$1
PROJECT=$2
REVISION=$3
DO=$4

PROJECT_PATH="${GOPATH}/src/${PROJECT}"

# Do initial checkout if it doesn't exist
$(cd "$PROJECT_PATH" 2> /dev/null) || git clone ${REPO} ${PROJECT_PATH}
pushd "$PROJECT_PATH"
# Attempt to checkout the specified revision
git fetch --all
git checkout ${REVISION}
${DO}
