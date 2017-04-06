// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"fmt"
	"io"
	"os"
)

// Ensure the directory exists or create it if needed.
func EnsureDir(dir string, mode os.FileMode) error {
	if fileOptions, err := os.Stat(dir); os.IsNotExist(err) {
		if errMake := os.MkdirAll(dir, mode); errMake != nil {
			return fmt.Errorf("Could not create directory %s. %v", dir, err)
		}
	} else if err != nil {
		return fmt.Errorf("Error asserting directory %s: %v", dir, err)
	} else if !fileOptions.IsDir() {
		return fmt.Errorf("Path already exists as a file: %s", dir)
	}
	return nil
}

// Check whether the provided directory exists
func IsDir(directory string) bool {
	fileInfo, err := os.Stat(directory)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

// copied from http://stackoverflow.com/questions/21060945/simple-way-to-copy-a-file-in-golang
// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Otherwise, attempt to create a hard link
// between the two files. If that fail, copy the file contents from src to dst.
func CopyFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	// NOTE: [ben] we do not want to create a hard link currently
	// if err = os.Link(src, dst); err == nil {
	//     return
	// }
	err = copyFileContents(src, dst)
	return
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all its contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	// TODO: [ben] this blocks, so copy should be put in go-routine
	err = out.Sync()
	return
}
