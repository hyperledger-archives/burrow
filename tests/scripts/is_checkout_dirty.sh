#!/usr/bin/env bash
# Gives us a non-zero exit code if there are tracked or untracked changes in the working
# directory
exit $(git status --porcelain | wc -l)
