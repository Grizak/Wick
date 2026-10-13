.PHONY: release snapshot build clean check package

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