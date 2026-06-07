package win

type Win_Arm64 struct{}

func (t Win_Arm64) Conv() string {
	return "windows/arm64"
}

func (t Win_Arm64) Triple() string {
	return "aarch64-pc-windows-msvc"
}

func (t Win_Arm64) SysExit() string {
	return "@ExitProcess(i32 %code)"
}

func (t Win_Arm64) EntryPoint() string {
	return "mainCRTStartup"
}

func (t Win_Arm64) DataLayout() string {
	return "e-m:e-i8:8:32-i16:16:32-i32:32-i64:64-i128:128-n32:64-S128"
}
