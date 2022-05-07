package fuzzer

import (
	"fmt"
)

func Fuzz(data []byte) int {
	fmt.Println(data)
	return 0
}
