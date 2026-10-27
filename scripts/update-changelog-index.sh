#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
    echo "Usage: $0 <path-to-changelog-dir>" >&2
    exit 1
fi

dir="$1"

if [[ ! -d "$dir" ]]; then
    echo "Error: '$dir' is not a directory." >&2
    exit 1
fi

shopt -s nullglob
files=()
for f in "$dir"/*.md; do
    name="${f##*/}"
    [[ "$name" == "index.md" ]] && continue
    files+=("$name")
done

# Empty input produces an empty JSON array
printf '%s\n' "${files[@]+"${files[@]}"}" \
  | jq -R -s -c 'split("\n") | map(select(length > 0)) | map({file: ., title: (capture("(?<title>.*)\\.md$") | .title)})'