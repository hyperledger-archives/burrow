#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

# Gives the tag if we are exactly on a tag otherwise gives the tag last reachable tag with short commit hash as suffix
tagish="$(git describe --tags)"

# Drop the version
echo "${tagish#v}"
