package linker

/*
#include "linker.h"
*/
import "C"
import "fmt"

func internalLink(target string, args []string) error {
	var err error
	switch target {
	case "x86_64-pc-windows-msvc", "aarch64-pc-windows-msvc":
		err = linkCOFFBinary(args)
	case "x86_64-pc-linux-gnu", "aarch64-pc-linux-gnu":
		err = linkELFBinary(args)
	case "x86_64-apple-darwin", "aarch64-apple-darwin":
		err = linkMachOBinary(args)
	default:
		err = fmt.Errorf("Unsupported target")
	}
	return err
}

func linkCOFFBinary(args []string) error {
	return nil
}

func linkELFBinary(args []string) error {
	return nil
}

func linkMachOBinary(args []string) error {
	return nil
}
