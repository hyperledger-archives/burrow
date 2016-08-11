package test

import (
	"github.com/eris-ltd/eris-db/test/fixtures"
	"testing"
)

// Needs to be in a _test.go file to be picked up
func TestWrapper(runner func() int) int {
	ffs := fixtures.NewFileFixtures("Eris-DB")
	defer ffs.RemoveAll()

	err := initGlobalVariables(ffs)

	if err != nil {
		panic(err)
	}

	// start a node
	ready := make(chan error)
	go newNode(ready)
	err = <-ready

	if err != nil {
		panic(err)
	}

	return runner()
}

func DebugMain() {
	t := &testing.T{}
	TestWrapper(func() int {
		testBroadcastTx(t, "HTTP")
		return 0
	})
}

func Successor(x int) int {
	return x + 1
}
