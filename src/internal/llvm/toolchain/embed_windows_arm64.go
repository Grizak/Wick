//go:build windows && arm64

package toolchain

import _ "embed"

//go:embed bin/windows_arm64/lld-link.exe
var lldBinary []byte

//go:embed bin/windows_arm64/versions.json
var versionsJSON []byte

//go:embed bin/windows_amd64/kernel32.lib
var kernel32Amd64Lib []byte

//go:embed bin/windows_arm64/kernel32.lib
var kernel32Arm64Lib []byte
