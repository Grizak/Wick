# Changelog

## [Unreleased]

## [v4.1.2] - 2026-10-27

- Update workflow to commit assets dir

## [v4.1.1] - 2026-10-27

- Updated the workflow to run the before hooks there instead of in goreleaser

## [v4.1.0] - 2026-10-27

- Added changelog command
- Made it read the changelog from inside the binary

## [v4.0.2] - 2026-10-26

Added changelog for this version and v4.0.1 because I forgot for that version and the release then subsequently failed

## [v4.0.1] - 2026-10-26

Removed "go mod tidy" from build ci because it broke goreleaser's git check

## [v4.0.0] - 2026-10-26

### Added

- New build backend with tinygo/llvm

### Not A Breaking Release!

- Old code will still continue functioning
- Old flags are still supported
- 4.0.0 just makes me feel good
