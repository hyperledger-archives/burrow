#!/usr/bin/env bash

set -e

function release {
    echo "Pushing dev release to $DOCKER_REPO_DEV..."
    echo ${DOCKER_PASS_DEV} | docker login --username ${DOCKER_USER_DEV} ${CI_REGISTRY} ${DOCKER_HUB_DEV} --password-stdin
    docker tag ${DOCKER_REPO}:$(./scripts/local_version.sh) ${DOCKER_REPO_DEV}:$(./scripts/local_version.sh)
    docker push ${DOCKER_REPO_DEV}
}

# Only do a dev release outside pull requests
if [[ -n "$CIRCLE_PULL_REQUESTS" ]]; then
    echo "CIRCLE_PULL_REQUESTS is set to: '$CIRCLE_PULL_REQUEST' so not pushing dev release"
    exit 0
fi

release