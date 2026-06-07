package darwin

type Darwin_Arm64 struct{}

func (t Darwin_Arm64) Conv() string {
	return "darwin/arm64"
}

func (t Darwin_Arm64) Triple() string {
	return "aarch64-apple-darwin"
}

func (t Darwin_Arm64) SysExit() string {
	return "asm sideeffect \"svc #0x80\", \"{x16},{x0}\" (i64 1, i64 %code64)"
}

func (t Darwin_Arm64) EntryPoint() string {
	return "_main"
}

func (t Darwin_Arm64) DataLayout() string {
	return "e-m:e-i8:8:32-i16:16:32-i32:32-i64:64-i128:128-n32:64-S128"
}
