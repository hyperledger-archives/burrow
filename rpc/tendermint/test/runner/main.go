// +build integration

// Space above here matters
package main

import (
	"fmt"

	rpctest "github.com/eris-ltd/eris-db/rpc/tendermint/test"
	"github.com/eris-ltd/eris-db/util"
)

func main() {
	//fmt.Printf("%s", util.IsAddress("hello"))
	fmt.Printf("%s", util.IsAddress("hello"), rpctest.Successor(2))
	//defer os.Exit(0)
}
