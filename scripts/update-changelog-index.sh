#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
    echo "Usage: $0 <path-to-changelog>"
    exit 1
fi

dir="$1"

if [[ ! -f "$dir" ]]; then
    echo "Error: '$dir' does not exist."
    exit 1
fi

# Generate the index file as json and write it to stdout
jq -R -s -c 'split("\n") | map(select(length > 0)) | map({file: ., title: (capture("(?<title>.*)\\.md$") | .title)})' "$dir"