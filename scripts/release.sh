#!/usr/bin/env bash

version_regex="^v[0-9]+\.[0-9]+\.[0-9]+$"

set -e

function release {
    notes="NOTES.md"
    echo "Building and releasing $tag..."
    echo "Pushing docker image..."
    echo ${DOCKER_PASS} | docker login --username ${DOCKER_USER} --password-stdin
    docker push ${DOCKER_REPO}
    echo "Building and pushing binaries"
    [[ -e "$notes" ]] && goreleaser --release-notes "$notes" || goreleaser
}


# If passed argument try to use that as tag otherwise read from local repo
if [[ $1 ]]; then
    # Override mode, try to release this tag
    export tag=$1
else
    echo "Getting tag from HEAD which is $(git rev-parse HEAD)"
    export tag=$(git tag --points-at HEAD)
fi

if [[ ! ${tag} ]]; then
    echo "No tag so not releasing."
    exit 0
fi

# Only release semantic version syntax tags
if [[ ! ${tag} =~ ${version_regex} ]] ; then
    echo "Tag '$tag' does not match version regex '$version_regex' so not releasing."
    exit 0
fi

release