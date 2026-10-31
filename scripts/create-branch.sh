#!/usr/bin/env bash

# Check that we got enough arguments
if [ $# -lt 3 ]; then
  echo "Usage: $0 <type> <number> <work>"
  exit 1
fi

TYPE=$1
NUM=$2
WORK=$3
BRANCH="$TYPE/$NUM/$WORK"

gh issue develop $NUM --name $BRANCH --base dev --checkout
git push -u origin $BRANCH
gh pr create --fill --draft --title "$TYPE: $WORK (#$NUM)" --body "\n\nCloses #$NUM."