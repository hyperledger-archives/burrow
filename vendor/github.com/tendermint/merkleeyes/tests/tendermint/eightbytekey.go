package main

import (
	"crypto/rand"
	"fmt"
)

func main() {
	buf := make([]byte, 8)
	rand.Read(buf)
	fmt.Printf("%x\n", buf)
}
