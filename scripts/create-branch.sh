#!/usr/bin/env bash
set -euo pipefail

if [ $# -lt 3 ]; then
  echo "Usage: $0 <type> <number> <work>" >&2
  exit 1
fi

TYPE=$1
NUM=$2
WORK=$3
MILESTONE=${4:-}
BRANCH="$TYPE/$NUM/$WORK"

if ! [[ "$NUM" =~ ^[0-9]+$ ]]; then
  echo "Error: issue number must be numeric" >&2
  exit 1
fi

if ! git check-ref-format --branch "$BRANCH" >/dev/null 2>&1; then
  echo "Error: '$BRANCH' is not a valid branch name" >&2
  exit 1
fi

MILESTONE_ARGS=()
if [ -n "$MILESTONE" ]; then
  MILESTONE_ARGS=(--milestone "$MILESTONE")
  # Set the milestone on the issue
  gh issue edit "$NUM" --milestone "$MILESTONE"
fi

gh issue develop "$NUM" --name "$BRANCH" --base dev --checkout

# PRs need at least one commit that differs from the base branch
git commit --allow-empty -m "chore: start work on #$NUM"
git push -u origin "$BRANCH"

# ${arr[@]+"${arr[@]}"} avoids an "unbound variable" error on empty arrays with set -u (bash < 4.4)
gh pr create --draft --title "$TYPE: $WORK (#$NUM)" --body "Closes #$NUM." \
  ${MILESTONE_ARGS[@]+"${MILESTONE_ARGS[@]}"}