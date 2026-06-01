//go:build windows && amd64

package tools

import _ "embed"

//go:embed bin/windows_amd64/llc.exe
var llcBinary []byte

//go:embed bin/windows_amd64/lld-link.exe
var lldBinary []byte

//go:embed bin/windows_amd64/kernel32.lib
var kernel32Lib []byte

//go:embed bin/windows_amd64/versions.json
var versionsJSON []byte
