package types

var TargetTriples = map[string]string{
	"linux/amd64":   "x86_64-pc-linux-gnu",
	"linux/arm64":   "aarch64-pc-linux-gnu",
	"darwin/amd64":  "x86_64-apple-darwin",
	"darwin/arm64":  "aarch64-apple-darwin",
	"windows/amd64": "x86_64-pc-windows-msvc",
	"windows/arm64": "aarch64-pc-windows-msvc",
}
