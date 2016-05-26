// Copyright 2015, 2016 Eris Industries (UK) Ltd.
// This file is part of Eris-RT

// Eris-RT is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Eris-RT is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Eris-RT.  If not, see <http://www.gnu.org/licenses/>.

package util

import (
  "fmt"
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
