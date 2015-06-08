// +build windows

package files

import (
	"fmt"
	"os"
)

// TODO finish up.
func Rename(oldname, newname string) error {

	// Some extra fluff here.
	if fs, err := os.Stat(newname); !os.IsNotExist(err) {
		if fs.Mode().IsRegular() && isWritable(fs.Mode().Perm()) {
			errRM := os.Remove(newname)
			if errRM != nil {
				return errRM
			}
		} else {
			return fmt.Errorf("Target exists and cannot be over-written (is a directory or read-only file): " + newname)
		}
	}
	errRN := os.Rename(oldname, newname)
	if errRN != nil {
		return errRN
	}

	return nil
}
