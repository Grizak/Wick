//go:build darwin && arm64

package tools

import _ "embed"

//go:embed bin/darwin_arm64/llc
var llcBinary []byte

//go:embed bin/darwin_arm64/ld64.lld
var lldBinary []byte

//go:embed bin/darwin_arm64/versions.json
var versionsJSON []byte
