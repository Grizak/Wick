#!/usr/bin/env bash

TARGET=$1
SOURCES=${@:2}

go build -ldflags "-X main.version=$(git describe --tags --abbrev=0)" -o $TARGET $SOURCES
