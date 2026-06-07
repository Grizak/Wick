package win

type Win_Amd64 struct{}

func (t Win_Amd64) Conv() string {
	return "windows/amd64"
}

func (t Win_Amd64) Triple() string {
	return "x86_64-pc-windows-msvc"
}

func (t Win_Amd64) SysExit() string {
	return "@ExitProcess(i32 %code)"
}

func (t Win_Amd64) EntryPoint() string {
	return "mainCRTStartup"
}

func (t Win_Amd64) DataLayout() string {
	return "e-m:w-p270:32:32-p271:32:32-p272:64:64-i64:64-f80:128-n8:16:32:64-S128"
}
