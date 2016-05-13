// Copyright 2015, 2016 Eris Industries (UK) Ltd.
// This file is part of Eris-RT

// Eris-RT is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Eris-RT is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Eris-RT.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"github.com/eris-ltd/eris-db/cmd"
)

func main() {
  commands.Execute()
}

// package main
//
// import (
// 	"fmt"
// 	edb "github.com/eris-ltd/eris-db/erisdb"
// 	"os"
// )
//
// // TODO the input stuff.
// func main() {
// 	args := os.Args[1:]
// 	var baseDir string
// 	var inProc bool
// 	if len(args) > 0 {
// 		baseDir = args[0]
// 		if len(args) > 1 {
// 			if args[1] == "inproc" {
// 				inProc = true
// 			}
// 		}
// 	} else {
// 		baseDir = os.Getenv("HOME") + "/.erisdb"
// 	}
//
// 	proc, errSt := edb.ServeErisDB(baseDir, inProc)
// 	if errSt != nil {
// 		panic(errSt.Error())
// 	}
// 	errSe := proc.Start()
// 	if errSe != nil {
// 		panic(errSe.Error())
// 	}
// 	// TODO For now.
// 	fmt.Println("DONTMINDME55891")
// 	<-proc.StopEventChannel()
// }
