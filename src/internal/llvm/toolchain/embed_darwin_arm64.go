//go:build darwin && arm64

package toolchain

import _ "embed"

//go:embed bin/darwin_arm64/ld64.lld
var lldBinary []byte

//go:embed bin/windows_amd64/kernel32.lib
var kernel32Amd64Lib []byte

//go:embed bin/windows_arm64/kernel32.lib
var kernel32Arm64Lib []byte

//go:embed bin/darwin_arm64/versions.json
var versionsJSON []byte
