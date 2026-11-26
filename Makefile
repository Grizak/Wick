.PHONY: release snapshot build clean check package dev test assets build-runtime configure-runtime

DEV_OUT ?= ./dist/wick
ROOT := $(CURDIR)
GOOS_HOST := $(shell go env GOOS)
GOARCH_HOST := $(shell go env GOARCH)
LLVM_INCLUDE := $(ROOT)/src/internal/llvm/toolchain/include/$(GOOS_HOST)_$(GOARCH_HOST)
MUSL_TOOLCHAIN := $(ROOT)/toolchains/musl/linux_amd64
MUSL_CC := $(MUSL_TOOLCHAIN)/bin/x86_64-linux-musl-gcc
MUSL_CXX := $(MUSL_TOOLCHAIN)/bin/x86_64-linux-musl-g++
PREFIX := $(ROOT)/install/runtime-$(GOOS_HOST)-$(GOARCH_HOST)

# Lazily evaluated, so the script only runs for targets that need it
LDFLAGS_HOST = $(shell $(ROOT)/scripts/llvm-ldflags.sh $(GOOS_HOST)_$(GOARCH_HOST))

dev: build-runtime assets
	CC=$(MUSL_CC) \
	CXX=$(MUSL_CXX) \
	CGO_ENABLED=1 \
	CGO_CFLAGS="-I$(LLVM_INCLUDE) -ffunction-sections -fdata-sections -flto" \
	CGO_CXXFLAGS="-I$(LLVM_INCLUDE) -std=c++17 -ffunction-sections -fdata-sections -flto" \
	CGO_LDFLAGS="$(LDFLAGS_HOST)" \
	go build -tags "byollvm llvm20" \
	-ldflags "-X main.version=dev -linkmode=external -extldflags=-static" \
	-o "$(DEV_OUT)" ./src/cmd/wick

TEST_FLAGS ?= -v ./src/...

test: build-runtime assets
	CC=gcc \
	CXX=g++ \
	CGO_ENABLED=1 \
	CGO_CFLAGS="-I$(LLVM_INCLUDE) -ffunction-sections -fdata-sections -flto" \
	CGO_CXXFLAGS="-I$(LLVM_INCLUDE) -std=c++17 -ffunction-sections -fdata-sections -flto" \
	CGO_LDFLAGS="$(LDFLAGS_HOST)" \
	go test -tags "byollvm llvm20" \
	 $(TEST_FLAGS)

GORELEASER_FLAGS ?=

release:
	goreleaser release --clean $(GORELEASER_FLAGS)

snapshot:
	goreleaser release --snapshot --clean $(GORELEASER_FLAGS)

check:
	goreleaser check $(GORELEASER_FLAGS)

clean:
	rm -rf dist build runtime/build

package:
	goreleaser release --skip=publish --clean $(GORELEASER_FLAGS)

setup:
	git config core.hooksPath .githooks

assets:
	@mkdir -p $(dir $(DEV_OUT))
	go run ./scripts/prepare-assets.go

build-runtime: configure-runtime
	@cmake --build runtime/build
	@cmake --install runtime/build

configure-runtime:
	@cmake -S runtime -B runtime/build -G Ninja \
		-DCMAKE_C_COMPILER=$(MUSL_CC) \
		-DCMAKE_CXX_COMPILER=$(MUSL_CXX) \
		-DCMAKE_C_COMPILER_LAUNCHER=sccache \
		-DCMAKE_CXX_COMPILER_LAUNCHER=sccache \
		-DCMAKE_BUILD_TYPE=Release \
		-DCMAKE_INSTALL_PREFIX="$(PREFIX)"
