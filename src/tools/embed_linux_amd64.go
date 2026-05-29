//go:build linux && amd64

package tools

import _ "embed"

//go:embed bin/linux_amd64/llc
var llcBinary []byte

//go:embed bin/linux_amd64/ld.lld
var lldBinary []byte

//go:embed bin/linux_amd64/versions.json
var versionsJSON []byte
