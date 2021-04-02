package testutil

import (
	"fmt"
	"io/ioutil"
	"os"
)

const scratchDir = "test_scratch"

func EnterTestDirectory() (testDir string, cleanup func()) {
	var err error
	testDir, err = ioutil.TempDir("", scratchDir)
	if err != nil {
		panic(fmt.Errorf("could not make temp dir for integration tests: %v", err))
	}
	// If you need to inspectdirs
	//testDir := scratchDir
	os.RemoveAll(testDir)
	os.MkdirAll(testDir, 0777)
	os.Chdir(testDir)
	os.MkdirAll("config", 0777)
	return testDir, func() { os.RemoveAll(testDir) }
}
