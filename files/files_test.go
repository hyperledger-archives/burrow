package files

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"testing"
)

var tempFolder = os.Getenv("TEMP")
var fileData = []byte("aaaaaaaaaaaaaaaaa")

func TestWrite(t *testing.T) {
	fileName := "testfile"
	err := write(fileName)
	if err != nil {
		t.Fatal(err)
	}
	remove(fileName)
	if !checkGone(fileName){
		t.Error(fmt.Errorf("File not removed: " + fileName))
	}
}

func TestRead(t *testing.T) {
	fileName := "testfile"
	err := write(fileName)
	if err != nil {
		t.Fatal(err)
	}
	err2 := readAndCheck(fileName, fileData)
	if err2 != nil {
		remove(fileName)
		t.Fatal(err)
	}
	remove(fileName)
	if !checkGone(fileName){
		t.Error(fmt.Errorf("File not removed: " + fileName))
	}
}

func TestRename(t *testing.T) {
	fileName0 := "file0"
	fileName1 := "file1"
	err := write(fileName0)
	if err != nil {
		t.Fatal(err)
	}
	err2 := rename(fileName0,fileName1)
	if err2 != nil {
		t.Fatal(err2)
	}
	err3 := readAndCheck(fileName1, fileData)
	if err3 != nil {
		remove(fileName1)
		t.Fatal(err)
	}
	remove(fileName1)
	
	if !checkGone(fileName0){
		t.Error(fmt.Errorf("File not removed: " + fileName0))
	}
	if !checkGone(fileName1){
		t.Error(fmt.Errorf("File not removed: " + fileName1))
	}
}

func getName(name string) string {
	return path.Join(tempFolder, name)
}

func write(fileName string) error {
	return WriteFile(getName(fileName), fileData, FILE_RW)
}

func readAndCheck(fileName string, btsIn []byte) error {
	bts, err := ReadFile(getName(fileName))
	if err != nil {
		return err
	}
	if !bytes.Equal(bts,btsIn) {
		return fmt.Errorf("Failed to read file data. Written: %s, Read: %s\n", string(fileData), string(bts))
	}
	return nil
}

func remove(fileName string) error {
	return os.Remove(getName(fileName))
}

func rename(fileName0, fileName1 string) error {
	return Rename(getName(fileName0), getName(fileName1))
}

func checkGone(fileName string) bool {
	name := getName(fileName)
	_ , err := os.Stat(name) 
	return os.IsNotExist(err)
}