#!/usr/bin/env sh
set -eu

BASE="${GITHUB_BASE_SHA:-$(git rev-parse HEAD~1)}"
HEAD="${GITHUB_SHA:-$(git rev-parse HEAD)}"

git diff --unified=0 "$BASE" "$HEAD" \
  | /bin/keysweep-scanner