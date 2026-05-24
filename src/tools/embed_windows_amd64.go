//go:build windows && amd64

package tools

import _ "embed"

//go:embed bin/windows_amd64/nasm
var nasmBinary []byte

//go:embed bin/windows_amd64/lld
var lldBinary []byte
