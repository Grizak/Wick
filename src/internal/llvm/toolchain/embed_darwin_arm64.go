//go:build darwin && arm64

package toolchain

import _ "embed"

//go:embed bin/darwin_arm64/llc
var llcBinary []byte

//go:embed bin/darwin_arm64/ld64.lld
var lldBinary []byte

//go:embed bin/windows_amd64/kernel32.lib
var kernel32Lib []byte

//go:embed bin/darwin_arm64/versions.json
var versionsJSON []byte

//go:embed bin/darwin_arm64/opt
var optBinary []byte
