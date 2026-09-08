#!/usr/bin/env bash

RELEASE=$WICK_RELEASE
TARGET=$1
SOURCES=${@:2}

printf "%s " "Building a" 
if [[ -z $RELEASE ]]; then
 printf "development"
else
 printf "release"
fi
echo " build of WICK for $GOOS/$GOARCH"

if [ -z $RELEASE ]; then # If !RELEASE
  go build -ldflags "-X main.version=$(git describe --tags --abbrev=0)" -o $TARGET $SOURCES
  echo "Done"
else
  go build -ldflags "-X main.version=$(git describe --tags --abbrev=0) -s -w" -o $TARGET $SOURCES # Build without debug symbols (-s -w)
  echo "Done"
fi
