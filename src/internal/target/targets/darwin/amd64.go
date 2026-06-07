package darwin

type Darwin_Amd64 struct{}

func (t Darwin_Amd64) Conv() string {
	return "darwin/amd64"
}

func (t Darwin_Amd64) Triple() string {
	return "aarch64-apple-darwin"
}

func (t Darwin_Amd64) SysExit() string {
	return "asm sideeffect \"syscall\", \"{rax},{rdi}\" (i64 0x2000001, i64 %code64)"
}

func (t Darwin_Amd64) EntryPoint() string {
	return "_main"
}

func (t Darwin_Amd64) DataLayout() string {
	return "e-m:o-p270:32:32-p271:32:32-p272:64:64-i64:64-f80:128-n8:16:32:64-S128"
}
