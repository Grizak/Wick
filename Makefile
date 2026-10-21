.PHONY: release snapshot build clean check package dev

DEV_OUT ?= ./dist/wick
GOOS_HOST := $(shell go env GOOS)
GOARCH_HOST := $(shell go env GOARCH)

LLVM_TOOLCHAIN := $(CURDIR)/src/internal/llvm/toolchain
LLVM_INCLUDE := $(LLVM_TOOLCHAIN)/include/$(GOOS_HOST)_$(GOARCH_HOST)
LLVM_LIBDIR  := $(LLVM_TOOLCHAIN)/lib/$(GOOS_HOST)_$(GOARCH_HOST)

ifeq ($(GOOS_HOST),windows)
LLVM_LIB := $(LLVM_LIBDIR)/LLVM.lib
else
LLVM_LIB := $(LLVM_LIBDIR)/libLLVM.a
endif

dev:
	mkdir -p $(dir $(DEV_OUT))
	mkdir -p src/internal/assets
	cp LICENSE src/internal/assets/LICENSE
	CC=gcc CXX=g++ \
	CGO_ENABLED=1 \
	CGO_CFLAGS="-I$(LLVM_INCLUDE)" \
	CGO_LDFLAGS="$(LLVM_LIB)" \
	go build -tags "byollvm llvm20" -ldflags "-X main.version=dev" -o "$(DEV_OUT)" ./src/cmd/wick

release:
	goreleaser release --clean

snapshot:
	goreleaser release --snapshot --clean

build:
	goreleaser build --snapshot --clean

check:
	goreleaser check

clean:
	rm -rf dist build

package:
	goreleaser release --skip=publish --clean