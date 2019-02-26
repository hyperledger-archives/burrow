#!/usr/bin/env bash

set -e

DOCKER_REPO=${DOCKER_REPO:-"hyperledger/burrow"}
DOCKER_REPO_DEV=${DOCKER_REPO_DEV:-"quay.io/monax/burrow"}

function release {
    echo "Pushing dev release to $DOCKER_REPO_DEV..."
    echo ${DOCKER_PASS_DEV} | docker login --username ${DOCKER_USER_DEV} ${CI_REGISTRY} ${DOCKER_HUB_DEV} --password-stdin
    docker tag ${DOCKER_REPO}:$(./scripts/local_version.sh) ${DOCKER_REPO_DEV}:$(./scripts/local_version.sh)
    docker push ${DOCKER_REPO_DEV}
}

release