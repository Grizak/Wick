package linker

/*
#include <stdlib.h>
#include "lld_bridge.h"
*/
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/Grizak/Wick/src/internal/target"
)

func internalLink(targetTriple string, args []string) error {
	args, err := pushTarget(target.BinaryFormat[targetTriple], args)
	if err != nil {
		return err
	}
	cargs := make([]*C.char, len(args))
	for i, arg := range args {
		cargs[i] = C.CString(arg)
		defer C.free(unsafe.Pointer(cargs[i]))
	}

	var argv **C.char
	if len(cargs) > 0 {
		argv = (**C.char)(unsafe.Pointer(&cargs[0]))
	}

	code := C.Run_lld(C.int(len(cargs)), argv)
	if code != 0 {
		return fmt.Errorf("lld failed with code %d", int(code))
	}
	return nil
}

func ExternalLink(targetTriple string, args []string) error {
	return internalLink(targetTriple, args)
}
