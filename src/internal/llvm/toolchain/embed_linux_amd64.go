//go:build linux && amd64

package toolchain

import _ "embed"

//go:embed bin/linux_amd64/ld.lld
var lldBinary []byte

//go:embed bin/windows_amd64/kernel32.lib
var kernel32Amd64Lib []byte

//go:embed bin/windows_arm64/kernel32.lib
var kernel32Arm64Lib []byte

//go:embed bin/linux_amd64/versions.json
var versionsJSON []byte
