package main

import (
	"fmt"
	edb "github.com/eris-ltd/eris-db/erisdb"
	"os"
)

// TODO the input stuff.
func main() {
	args := os.Args[1:]
	var baseDir string
	var inProc bool
	if len(args) > 0 {
		baseDir = args[0]
		if len(args) > 1 {
			if args[1] == "inproc" {
				inProc = true
			}
		}
	} else {
		baseDir = os.Getenv("HOME") + "/.erisdb"
	}

	proc, errSt := edb.ServeErisDB(baseDir, inProc)
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
