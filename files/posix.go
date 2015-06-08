// +build !windows

package files

import "os"

// Rename for linux and macs etc. Don't really care about the rest.
func Rename(oldname, newname string) error {
	return os.Rename(oldname, newname)
}