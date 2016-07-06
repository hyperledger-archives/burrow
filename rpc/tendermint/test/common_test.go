package rpctest

import (
	"testing"
	"github.com/eris-ltd/eris-db/test/fixtures"
	"fmt"
	"os"
)

// Needs to be in a _test.go file to be picked up
func TestMain(m *testing.M) {
	ffs := fixtures.NewFileFixtures("Eris-DB")
	defer ffs.RemoveAll()

	err := initGlobalVariables(ffs)

	if err != nil {
		panic(err)
	}

	saveNewPriv()

	// start a node

	ready := make(chan error)
	go newNode(ready)
	err = <-ready

	if err != nil {
		panic(err)
	}

	returnValue := m.Run()
	fmt.Println("foooooo", returnValue)

	defer os.Exit(returnValue)
}

