package common

import (
	"os"
	"path/filepath"
	"runtime"
)

var (
	// Convenience directories.
	MonaxRoot          = ResolveMonaxRoot()
	MonaxContainerRoot = "/home/monax/.monax"

	// Major directories.
	KeysPath    = filepath.Join(MonaxRoot, "keys")
	ScratchPath = filepath.Join(MonaxRoot, "scratch")
)

func HomeDir() string {
	if runtime.GOOS == "windows" {
		drive := os.Getenv("HOMEDRIVE")
		path := os.Getenv("HOMEPATH")
		if drive == "" || path == "" {
			return os.Getenv("USERPROFILE")
		}
		return drive + path
	} else {
		return os.Getenv("HOME")
	}
}

func ResolveMonaxRoot() string {
	var monax string
	if os.Getenv("MONAX") != "" {
		monax = os.Getenv("MONAX")
	} else {
		if runtime.GOOS == "windows" {
			home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
			if home == "" {
				home = os.Getenv("USERPROFILE")
			}
			monax = filepath.Join(home, ".monax")
		} else {
			monax = filepath.Join(HomeDir(), ".monax")
		}
	}
	return monax
}
