package main

import (
	"fmt"

	"github.com/hyperledger/burrow/project"
)

func main() {
	fmt.Println(project.History.CurrentVersion().String())
}
