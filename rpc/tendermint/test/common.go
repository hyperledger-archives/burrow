
// Space above here matters
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
