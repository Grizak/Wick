package utils

import (
	"fmt"
	"os"
)

var debug bool

func SetDebug() {
	debug = true
}

func Debug(msg string) {
	if debug {
		fmt.Fprintf(os.Stderr, "%s\n", msg)
	}
}
