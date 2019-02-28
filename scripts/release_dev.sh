#!/usr/bin/env bash

set -e

DOCKER_REPO=${DOCKER_REPO:-"hyperledger/burrow"}
DOCKER_REPO_DEV=${DOCKER_REPO_DEV:-"quay.io/monax/burrow"}

function release {
    echo "Pushing dev release to $DOCKER_REPO_DEV..."
    [[ -z "$DOCKER_PASS_DEV" ]] && echo "\$DOCKER_PASS_DEV must be set to release dev version" && exit 1
    version=$(./scripts/local_version.sh)
    echo ${DOCKER_PASS_DEV} | docker login --username ${DOCKER_USER_DEV} ${CI_REGISTRY} ${DOCKER_HUB_DEV} --password-stdin
    docker tag ${DOCKER_REPO}:${version} ${DOCKER_REPO_DEV}:${version}
    docker push ${DOCKER_REPO_DEV}
}


# Only do a dev release outside pull requests
if [[ -n "$CIRCLE_PULL_REQUESTS" ]]; then
    echo "CIRCLE_PULL_REQUESTS is set to: '$CIRCLE_PULL_REQUEST' so not pushing dev release"
    exit 0
fi

release