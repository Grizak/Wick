#!/usr/bin/env bash
# Splits CHANGELOG.md into release-notes/v<version>.md, one file per release.
set -euo pipefail

root=$(cd "$(dirname "$0")/.." && pwd)
src="$root/CHANGELOG.md"
out="$root/release-notes"

[ -f "$src" ] || { echo "no CHANGELOG.md at $src" >&2; exit 1; }

mkdir -p "$out"
# Clear old generated files so removed sections don't linger
rm -f "$out"/v*.md

awk -v out="$out" '
  /^## \[/ {
    if (file != "") close(file)
    ver = $0
    sub(/^## \[/, "", ver)
    sub(/\].*/, "", ver)
    if (ver == "Unreleased") { file = ""; next }
    if (ver !~ /^v/) ver = "v" ver
    file = out "/" ver ".md"
    printf "" > file
    next
  }
  file != "" { print > file }
' "$src"