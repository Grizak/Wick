package linux

type Linux_Arm64 struct{}

func (t Linux_Arm64) Conv() string {
	return "linux/arm64"
}

func (t Linux_Arm64) Triple() string {
	return "aarch64-pc-linux-gnu"
}

func (t Linux_Arm64) SysExit() string {
	return "asm sideeffect\"svc #0\", \"{x8},{x0}\" (i64 93, i64 %code64)"
}

func (t Linux_Arm64) EntryPoint() string {
	return "_start"
}

func (t Linux_Arm64) DataLayout() string {
	return "e-m:e-i8:8:32-i16:16:32-i32:32-i64:64-i128:128-n32:64-S128"
}
