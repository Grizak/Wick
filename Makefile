GO = go
BUILD_DIR ?= build
TARGET ?= wick
DARWIN_TARGET ?= $(TARGET)-darwin
WIN_TARGET ?= $(TARGET)-windows.exe
LINUX_TARGET ?= $(TARGET)-linux
TARGET_FULL = $(BUILD_DIR)/$(TARGET)
DARWIN_TARGET_FULL = $(BUILD_DIR)/$(DARWIN_TARGET)
WIN_TARGET_FULL = $(BUILD_DIR)/$(WIN_TARGET)
LINUX_TARGET_FULL = $(BUILD_DIR)/$(LINUX_TARGET)
SOURCES = ./src/cmd/wick/...

ASSETS_DIR = src/internal/assets
ASSETS_LICENSE = $(ASSETS_DIR)/LICENSE

.PHONY: current all clean art

current: prepare
	./build.sh $(TARGET_FULL) $(SOURCES)

all: $(DARWIN_TARGET_FULL) $(WIN_TARGET_FULL) $(LINUX_TARGET_FULL)
	@echo "Done"

prepare: $(ASSETS_LICENSE)
	mkdir -p $(BUILD_DIR)

art:
	rm -rf dist 2>/dev/null

clean: art
	rm -rf $(BUILD_DIR) 2>/dev/null

$(ASSETS_DIR)/%:
	cp $* $@

$(TARGET_FULL): prepare
	./build.sh $@ $(SOURCES)

$(DARWIN_TARGET_FULL): prepare
	GOOS=darwin GOARCH=arm64 WICK_RELEASE=yes ./build.sh $@ $(SOURCES)

$(WIN_TARGET_FULL): prepare
	GOOS=windows GOARCH=amd64 WICK_RELEASE=yes ./build.sh $@ $(SOURCES)

$(LINUX_TARGET_FULL): prepare
	GOOS=linux GOARCH=amd64 WICK_RELEASE=yes ./build.sh $@ $(SOURCES)