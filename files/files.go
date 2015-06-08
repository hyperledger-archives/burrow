
package files

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

// We don't concern ourselves with executable files here.
const (
	FILE_RW = os.FileMode(0666)
	FILE_W = os.FileMode(0x222) 
	FILE_R = os.FileMode(0444)
)

func isWritable(fm os.FileMode) bool {
	return fm & 2 == 2
}

func WriteFileRW(fileName string, data []byte) error {
	return WriteFile(fileName, data, FILE_RW)
}

func WriteFileReadOnly(fileName string, data []byte) error {
	return WriteFile(fileName, data, FILE_R)
}

func WriteFileWriteOnly(fileName string, data []byte) error {
	return WriteFile(fileName, data, FILE_W)
}

// WriteFile. Will do a 
func WriteFile(filename string, data []byte, perm os.FileMode) error {
	f, err := os.Create(filename)
	
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
		err2 = io.ErrShortWrite
		return err
	}
	
	return nil
}

func ReadFile(fileName string) ([]byte, error) {
	return ioutil.ReadFile(fileName)
}