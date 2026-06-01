GO = go
BUILD_DIR = build
TARGET ?= wick
TARGET_FULL = $(BUILD_DIR)/$(TARGET)
SOURCES = src/main.go

.PHONY: all clean art

all:
	mkdir -p $(BUILD_DIR)
	$(GO) build -ldflags "-X main.version=$(shell git describe --tags --abbrev=0)" -o $(TARGET_FULL) $(SOURCES)

art:
	rm *.ll *.o *.asm out 2>/dev/null

clean: art
	rm -rf $(BUILD_DIR) 2>/dev/null
