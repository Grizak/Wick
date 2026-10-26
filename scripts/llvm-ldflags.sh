#!/usr/bin/env bash
# Prints CGO_LDFLAGS for one platform, e.g. "linux_amd64", from the copied LLVM archives.
set -euo pipefail

platform=${1:?usage: llvm-ldflags.sh <platform>}
root=$(cd "$(dirname "$0")/.." && pwd)
dir="$root/src/internal/llvm/toolchain/lib/$platform"

if [ ! -d "$dir" ]; then
  echo "no LLVM libraries in $dir" >&2
  exit 1
fi

libs=()
for f in "$dir"/lib*.a; do
  base=$(basename "$f" .a)
  libs+=("-l${base#lib}")
done

if [ "${#libs[@]}" -eq 0 ]; then
  echo "no lib*.a archives in $dir" >&2
  exit 1
fi

case "$platform" in
  linux_*)
    echo "-L$dir -Wl,--start-group ${libs[*]} -Wl,--end-group -static-libstdc++ -static-libgcc -lstdc++ -lm -ldl -pthread"
    ;;
  windows_amd64)
    echo "-L$dir -Wl,--start-group ${libs[*]} -Wl,--end-group -lstdc++ -lpsapi -lshell32 -lole32 -luuid -ladvapi32"
    ;;
  windows_arm64)
    # Use -lc++ or -lstdc++, whichever the archive references (check with nm)
    echo "-L$dir -Wl,--start-group ${libs[*]} -Wl,--end-group -lc++ -lpsapi -lshell32 -lole32 -luuid -ladvapi32"
    ;;
  darwin_*)
    # ld64 has no --start-group. Order matters, so if you see undefined
    # symbols, add -Wl,-all_load here.
    echo "-L$dir ${libs[*]} -lc++ -lm"
    ;;
  *)
    echo "unknown platform: $platform" >&2
    exit 1
    ;;
esac