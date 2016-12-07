package util

import (
	"fmt"
	"os"
)

// Prints an error message to stderr and exits with status code 1
func Fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

