#!/usr/bin/env bash
# Gives us a non-zero exit code if there are tracked or untracked changes in the working
# directory
export stat=$(git status --porcelain)
[[ -z "$stat" ]] || (echo "Dirty checkout:" && echo "$stat" && exit 1)
