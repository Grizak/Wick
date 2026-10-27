#!/usr/bin/env bash

RELEASE=$WICK_RELEASE
TARGET=$1
SOURCES=${@:2}

export CGO_CFLAGS="-I$(llvm-config-20 --includedir)"
export CGO_LDFLAGS="$(llvm-config-20 --ldflags --libs --system-libs) -lstdc++"
export CGO_CPPFLAGS="-I$(llvm-config-20 --includedir)"
export CGO_CXXFLAGS="-std=c++17"
export CGO_ENABLED=1

if [ -z $RELEASE ]; then # If !RELEASE
  go build -ldflags "-X main.version=$(git describe --tags --abbrev=0)" -o $TARGET $SOURCES
else
  go build -ldflags "-X main.version=$(git describe --tags --abbrev=0) -s -w" -o $TARGET $SOURCES # Build without debug symbols (-s -w)
fi