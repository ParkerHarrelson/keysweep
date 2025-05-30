#!/usr/bin/env sh
set -eu

#
# 1. Read inputs (from action.yml → $INPUT_<NAME>)
#
USE_DEFAULT="${INPUT_USE_DEFAULT_RULES:-true}"
CUSTOM_PATH="${INPUT_CUSTOM_RULES_PATH:-}"
CUSTOM_URL="${INPUT_CUSTOM_RULES_URL:-}"
BASE_REF="${INPUT_BASE_REF:-}"
HEAD_SHA="${INPUT_HEAD_SHA:-}"

#
# 2. Determine commits for diff
#
if [ -n "$BASE_REF" ]; then
  git fetch origin "$BASE_REF":"$BASE_REF"
  BASE="$BASE_REF"
else
  BASE=$(git rev-parse HEAD~1)
fi

if [ -n "$HEAD_SHA" ]; then
  HEAD="$HEAD_SHA"
else
  HEAD=$(git rev-parse HEAD)
fi

#
# 3. Merge rule files into a single TOML
#
COMBINED="/tmp/combined.toml"
: > "$COMBINED"

# 3a) default rules
if [ "$USE_DEFAULT" = "true" ] && [ -f "/workspace/.gitleaks.toml" ]; then
  cat "/workspace/.gitleaks.toml" >> "$COMBINED"
fi

# 3b) custom in-repo
if [ -n "$CUSTOM_PATH" ] && [ -f "/workspace/$CUSTOM_PATH" ]; then
  cat "/workspace/$CUSTOM_PATH" >> "$COMBINED"

# 3c) custom remote
elif [ -n "$CUSTOM_URL" ]; then
  curl -fsSL "$CUSTOM_URL" >> "$COMBINED"
fi

 # 3d) pick the config we’ll feed to gitleaks
 if [ -s "$COMBINED" ]; then
   CONFIG="$COMBINED"
 else
   echo "❌ No gitleaks config found (default or custom)" >&2
   exit 1
 fi

# 3e) If someone piped data into STDIN (e.g. smoke test), skip the git-diff path
if [ ! -t 0 ]; then
  cat - | /bin/keysweep-scanner --config "$CONFIG"
  exit $?
fi

 #
 # 4. Run the PR diff through KeySweep scanner
 #
 git diff --unified=0 "$BASE" "$HEAD" \
   | /bin/keysweep-scanner --config "$CONFIG"

