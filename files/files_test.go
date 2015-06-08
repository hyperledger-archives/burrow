package files

import (
	"bytes"
	"os"
	"path"
	"testing"
	"github.com/stretchr/testify/assert"
)

var tempFolder = os.TempDir()
var fileData = []byte("aaaaaaaaaaaaaaaaa")

func TestWriteRemove(t *testing.T) {
	fileName := "testfile"
	write(t, fileName)
	remove(t, fileName)
}

func TestWriteReadRemove(t *testing.T) {
	fileName := "testfile"
	write(t, fileName)
	readAndCheck(t, fileName, fileData)
	remove(t, fileName)
}

func TestRenameRemove(t *testing.T) {
	fileName0 := "file0"
	fileName1 := "file1"
	write(t, fileName0)
	rename(t, fileName0, fileName1)
	readAndCheck(t, fileName1, fileData)
	remove(t, fileName1)
	checkGone(t, fileName0)
}

// Helpers

func getName(name string) string {
	return path.Join(tempFolder, name)
}

func write(t *testing.T, fileName string) {
	err := WriteFile(getName(fileName), fileData, FILE_RW)
	assert.NoError(t, err)
}

func readAndCheck(t *testing.T, fileName string, btsIn []byte) {
	bts, err := ReadFile(getName(fileName))
	assert.NoError(t, err)
	assert.True(t, bytes.Equal(bts,btsIn), "Failed to read file data. Written: %s, Read: %s\n", string(fileData), string(bts))
}

func remove(t *testing.T, fileName string) {
	err := os.Remove(getName(fileName))
	assert.NoError(t, err)
	checkGone(t, fileName)
}

func rename(t *testing.T, fileName0, fileName1 string) {
	assert.NoError(t, Rename(getName(fileName0), getName(fileName1)))
}

func checkGone(t *testing.T, fileName string) {
	name := getName(fileName)
	_ , err := os.Stat(name)
	assert.True(t, os.IsNotExist(err), "File not removed: " + name)
}