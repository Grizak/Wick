//go:build darwin && amd64

package tools

import _ "embed"

//go:embed bin/darwin_amd64/nasm
var nasmBinary []byte

//go:embed bin/darwin_amd64/lld
var lldBinary []byte
