package rpctest

import (
	"testing"
	"github.com/eris-ltd/eris-db/test/fixtures"
	"os"
)

// Needs to be in a _test.go file to be picked up
func TestMain(m *testing.M) {
	ffs := fixtures.NewFileFixtures()
	defer ffs.RemoveAll()

	initGlobalVariables(ffs)

	if ffs.Error != nil {
		panic(ffs.Error)
	}

	saveNewPriv()

	// start a node

	ready := make(chan struct{})
	go newNode(ready)
	<-ready

	returnValue := m.Run()


	os.Exit(returnValue)
}

