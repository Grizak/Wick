package linux

type Linux_Amd64 struct{}

func (t Linux_Amd64) Conv() string {
	return "linux/amd64"
}

func (t Linux_Amd64) Triple() string {
	return "x86_64-pc-linux-gnu"
}

func (t Linux_Amd64) SysExit() string {
	return "asm sideeffect\"syscall\", \"{rax},{rdi}\" (i64 60, i64 %code64)"
}

func (t Linux_Amd64) EntryPoint() string {
	return "_start"
}

func (t Linux_Amd64) DataLayout() string {
	return "e-m:e-p270:32:32-p271:32:32-p272:64:64-i64:64-f80:128-n8:16:32:64-S128"
}
