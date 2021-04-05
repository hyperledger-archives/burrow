package encoding

import (
	"math/big"
	"strings"

	"github.com/hyperledger/burrow/encoding/web3hex"
)

// Convert Burrow's ChainID to a *big.Int so it can be used as a nonce for Ethereum signing.
// For compatibility with Ethereum tooling this function first tries to interpret the ChainID as an integer encoded
// either as an eth-style  0x-prefixed hex string or a base 10 integer, falling back to interpreting the string's
// raw bytes as a big-endian integer
func GetEthChainID(chainID string) *big.Int {
	if strings.HasPrefix(chainID, "0x") {
		d := new(web3hex.Decoder)
		b := d.BigInt(chainID)
		if d.Err() == nil {
			return b
		}
	}
	b := new(big.Int)
	id, ok := b.SetString(chainID, 10)
	if ok {
		return id
	}
	return b.SetBytes([]byte(chainID))
}
