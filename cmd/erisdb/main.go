package main

import (
	"fmt"
	edb "github.com/eris-ltd/eris-db/erisdb"
	"os"

	_ "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/Sirupsen/logrus" // hack cuz godeps :(
	_ "github.com/tendermint/go-clist"                                               // godeps ...
)

// TODO the input stuff.
func main() {
	var baseDir string
	if len(os.Args) == 2 {
		baseDir = os.Args[1]
	} else {
		baseDir = os.Getenv("HOME") + "/.erisdb"
	}

	proc, errSt := edb.ServeErisDB(baseDir)
	if errSt != nil {
		panic(errSt.Error())
	}
	errSe := proc.Start()
	if errSe != nil {
		panic(errSe.Error())
	}
	// TODO For now.
	fmt.Println("DONTMINDME55891")
	<-proc.StopEventChannel()
}
