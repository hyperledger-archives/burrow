// +build integration

// Space above here matters
package test

import (
	"fmt"

	vm "github.com/eris-ltd/eris-db/manager/eris-mint/evm"
	rpc_core "github.com/eris-ltd/eris-db/rpc/tendermint/core"
	"github.com/eris-ltd/eris-db/test/fixtures"
)

// Needs to be referenced by a *_test.go file to be picked up
func TestWrapper(runner func() int) int {
	fmt.Println("Running with integration TestWrapper (rpc/tendermint/test/common.go)...")
	ffs := fixtures.NewFileFixtures("Eris-DB")

	defer ffs.RemoveAll()

	vm.SetDebug(true)
	err := initGlobalVariables(ffs)

	if err != nil {
		panic(err)
	}

	// start a node
	ready := make(chan error)
	server := make(chan *rpc_core.TendermintWebsocketServer)
	defer func() {
		// Shutdown -- make sure we don't hit a race on ffs.RemoveAll
		tmServer := <-server
		tmServer.Shutdown()
	}()

	go newNode(ready, server)
	err = <-ready

	if err != nil {
		panic(err)
	}

	return runner()
}

// This main function exists as a little convenience mechanism for running the
// delve debugger which doesn't work well from go test yet. In due course it can
// be removed, but it's flux between pull requests should be considered
// inconsequential, so feel free to insert your own code if you want to use it
// as an application entry point for delve debugging.
func DebugMain() {
	//t := &testing.T{}
	TestWrapper(func() int {
		//testNameReg(t, "JSONRPC")
		return 0
	})
}

func Successor(x int) int {
	return x + 1
}
