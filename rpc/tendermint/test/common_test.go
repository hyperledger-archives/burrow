package rpctest

import (
	"testing"
	"github.com/eris-ltd/eris-db/test/fixtures"
	"os"
	"fmt"
)

// Needs to be in a _test.go file to be picked up
func TestMain(m *testing.M) {
	ffs := fixtures.NewFileFixtures("Eris-DB")
	defer ffs.RemoveAll()
	fmt.Println("Defered!!")

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

