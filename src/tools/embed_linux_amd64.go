//go:build linux && amd64

package tools

import _ "embed"

//go:embed bin/linux_amd64/llc
var llcBinary []byte

//go:embed bin/linux_amd64/ld.lld
var lldBinary []byte

//go:embed bin/windows_amd64/kernel32.lib
var kernel32Lib []byte

//go:embed bin/linux_amd64/versions.json
var versionsJSON []byte
