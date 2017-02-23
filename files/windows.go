// +build windows

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
