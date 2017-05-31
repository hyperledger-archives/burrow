// Cross-platform file utils.
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
	"io/ioutil"
	"os"
)

// We don't concern ourselves with executable files here.
const (
	FILE_RW = os.FileMode(0666)
	FILE_W  = os.FileMode(0222)
	FILE_R  = os.FileMode(0444)
)

func IsWritable(fm os.FileMode) bool {
	return fm&2 == 2
}

// Write a file that has both read and write flags set.
func WriteFileRW(fileName string, data []byte) error {
	return WriteFile(fileName, data, FILE_RW)
}

// Write file with the read-only flag set.
func WriteFileReadOnly(fileName string, data []byte) error {
	return WriteFile(fileName, data, FILE_R)
}

// Write file with the write-only flag set.
func WriteFileWriteOnly(fileName string, data []byte) error {
	return WriteFile(fileName, data, FILE_W)
}

// WriteFile.
func WriteFile(fileName string, data []byte, perm os.FileMode) error {
	f, err := os.Create(fileName)
	if err != nil {
		fmt.Println("ERROR OPENING: " + err.Error())
		return err
	}
	defer f.Close()
	n, err2 := f.Write(data)
	if err2 != nil {
		fmt.Println("ERROR WRITING: " + err.Error())
		return err
	}
	if err2 == nil && n < len(data) {
		return err
	}

	return nil
}

// Does the file with the given name exist?
func FileExists(fileName string) bool {
	_, err := os.Stat(fileName)
	return !os.IsNotExist(err)
}

func IsRegular(fileName string) bool {
	fs, err := os.Stat(fileName)
	if err != nil {
		return false
	}
	return fs.Mode().IsRegular()
}

func WriteAndBackup(fileName string, data []byte) error {
	fs, err := os.Stat(fileName)
	fmt.Println("Write and backup")
	if err != nil {
		if os.IsNotExist(err) {
			WriteFileRW(fileName, data)
			return nil
		}
		return err
	}
	if !fs.Mode().IsRegular() {
		return fmt.Errorf("Not a regular file: " + fileName)
	}
	backupName := fileName + ".bak"
	fs, err = os.Stat(backupName)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		// We only work with regular files.
		if !fs.Mode().IsRegular() {
			return fmt.Errorf(backupName + " is not a regular file.")
		}
		errR := os.Remove(backupName)
		if errR != nil {
			return errR
		}
	}
	// Backup file should now be gone.
	// Read from original file.
	bts, errR := ReadFile(fileName)
	if errR != nil {
		return errR
	}
	// Write it to the backup.
	errW := WriteFileRW(backupName, bts)
	if errW != nil {
		return errW
	}
	// Write new bytes to original.
	return WriteFileRW(fileName, data)
}

func ReadFile(fileName string) ([]byte, error) {
	return ioutil.ReadFile(fileName)
}
