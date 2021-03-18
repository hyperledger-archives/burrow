package web3hex

import (
	bin "encoding/binary"
	"math/big"
	"strings"

	"github.com/hyperledger/burrow/crypto"
	"github.com/tmthrgd/go-hex"
)

type encoder struct {
}

var Encoder = new(encoder)

func (e *encoder) Bytes(bs []byte) string {
	return "0x" + hex.EncodeToString(bs)
}

func (e *encoder) BytesTrim(bs []byte) string {
	if len(bs) == 0 {
		return ""
	}
	str := hex.EncodeToString(bs)
	// Ethereum expects leading zeros to be removed from RLP encodings (SMH)
	str = strings.TrimLeft(str, "0")
	if len(str) == 0 {
		// Special case for zero
		return "0x0"
	}
	return "0x" + str
}

func (e *encoder) BigInt(x *big.Int) string {
	return e.BytesTrim(x.Bytes())
}

func (e *encoder) Uint64OmitEmpty(x uint64) string {
	if x == 0 {
		return ""
	}
	return e.Uint64(x)
}

func (e *encoder) Uint64(x uint64) string {
	bs := make([]byte, 8)
	bin.BigEndian.PutUint64(bs, x)
	return e.BytesTrim(bs)
}

func (e *encoder) Address(address crypto.Address) string {
	return e.BytesTrim(address.Bytes())
}
