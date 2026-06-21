#!/usr/bin/env bash

RELEASE=$WICK_RELEASE
TARGET=$1
SOURCES=${@:2}

if [ -z $RELEASE ]; then # If !RELEASE
  go build -ldflags "-X main.version=$(git describe --tags --abbrev=0)" -o $TARGET $SOURCES
else
  go build -ldflags "-X main.version=$(git describe --tags --abbrev=0) -s -w" -o $TARGET $SOURCES # Build without debug symbols (-s -w)
fi