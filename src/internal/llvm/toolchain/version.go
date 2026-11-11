package toolchain

/*
#include "llvm/Config/llvm-config.h"
*/
import "C"

func LLVMVersion() string {
	return C.LLVM_VERSION_STRING
}
