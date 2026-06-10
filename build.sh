#!/usr/bin/env bash

RELEASE=$WICK_RELEASE
TARGET=$1
SOURCES=${@:2}

if [ -z $RELEASE ]; then
  go build -ldflags "-X main.version=$(git describe --tags --abbrev=0)" -o $TARGET $SOURCES
else
  go build -ldflags "-X main.version=$(git describe --tags --abbrev=0) -s -w" -o $TARGET $SOURCES
fi