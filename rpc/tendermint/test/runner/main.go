package main

import (
	"fmt"

	"github.com/eris-ltd/eris-db/util"
	"github.com/eris-ltd/eris-db/rpc/tendermint/test"
)

func main() {
	fmt.Printf("%s", util.IsAddress("hello"), rpctest.Successor(2))
	//defer os.Exit(0)
}