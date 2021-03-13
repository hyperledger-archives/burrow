package crypto

import (
	"fmt"
	"math/big"
	"testing"
)

func TestGetEthChainID(t *testing.T) {
	chainIDString := "BurrowChain_FAB3C1-AB0FD1"
	chainID := GetEthChainID(chainIDString)
	b := new(big.Int).SetBytes([]byte(chainIDString))
	fmt.Println(b)
	fmt.Println(chainID)
	fmt.Printf("%X", chainID.Bytes())
}
