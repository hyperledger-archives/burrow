#!/usr/bin/env bash

set -e

# Wait a second so we don't see ephemeral file changes
sleep 1

# Don't tag if there is a dirty working dir
if ! git diff-index --quiet HEAD  ; then
    echo "Warning there appears to be uncommitted changes in the working directory:"
    git diff-index HEAD
    echo
    echo "Please commit them or stash them before tagging a release."
    echo
fi

version=v$(go run ./project/cmd/version/main.go)
notes=$(go run ./project/cmd/notes/main.go)

echo "This command will tag the current commit $(git rev-parse --short HEAD) as version $version"
echo "defined programmatically in project/releases.go with release notes:"
echo
echo "$notes" | sed 's/^/> /'
echo
echo "It will then push the version tag to origin."
echo
read -p "Do you want to continue? [Y\n]: " -r
# Just hitting return defaults to continuing
[[ $REPLY ]] && [[ ! $REPLY =~ ^[Yy]$ ]] && echo && exit 0
echo

# Create tag
echo "Tagging version $version with message:"
echo ""
echo "$notes"
echo ""
echo "$notes" | git tag -s -a ${version} -F-

# Push tag
git push origin ${version}

