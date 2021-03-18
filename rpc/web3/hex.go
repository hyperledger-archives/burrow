package web3

import (
	bin "encoding/binary"
	"fmt"
	"math/big"
	"strings"

	"github.com/hyperledger/burrow/crypto"
	"github.com/tmthrgd/go-hex"
)

type HexDecoder struct {
	error
	must bool
}

func (d *HexDecoder) Must() *HexDecoder {
	return &HexDecoder{must: true}
}

func (d *HexDecoder) Err() error {
	return d.error
}

func (d *HexDecoder) pushErr(err error) {
	if d.must {
		panic(err)
	}
	if d.error == nil {
		d.error = err
	}
}

func (d *HexDecoder) Bytes(hs string) []byte {
	hexString := strings.TrimPrefix(hs, "0x")
	// Ethereum returns odd-length hexString strings when it removes leading zeros
	if len(hexString)%2 == 1 {
		hexString = "0" + hexString
	}
	bs, err := hex.DecodeString(hexString)
	if err != nil {
		d.pushErr(fmt.Errorf("could not decode bytes from '%s': %w", hs, err))
	}
	return bs
}

func (d *HexDecoder) Address(hs string) crypto.Address {
	if hs == "" {
		return crypto.Address{}
	}
	address, err := crypto.AddressFromBytes(d.Bytes(hs))
	if err != nil {
		d.pushErr(fmt.Errorf("could not decode address from '%s': %w", hs, err))
	}
	return address
}

func (d *HexDecoder) BigInt(hs string) *big.Int {
	return new(big.Int).SetBytes(d.Bytes(hs))
}

func (d *HexDecoder) Uint64(hs string) uint64 {
	bi := d.BigInt(hs)
	if !bi.IsUint64() {
		d.pushErr(fmt.Errorf("%v is not uint64", bi))
	}
	return bi.Uint64()
}

func (d *HexDecoder) Int64(hs string) int64 {
	bi := d.BigInt(hs)
	if !bi.IsInt64() {
		d.pushErr(fmt.Errorf("%v is not int64", bi))
	}
	return bi.Int64()
}

type hexEncoder struct {
}

var HexEncoder = new(hexEncoder)

func (e *hexEncoder) Bytes(bs []byte) string {
	return "0x" + hex.EncodeToString(bs)
}

func (e *hexEncoder) BytesTrim(bs []byte) string {
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

func (e *hexEncoder) BigInt(x *big.Int) string {
	return e.BytesTrim(x.Bytes())
}

func (e *hexEncoder) Uint64OmitEmpty(x uint64) string {
	if x == 0 {
		return ""
	}
	return e.Uint64(x)
}

func (e *hexEncoder) Uint64(x uint64) string {
	bs := make([]byte, 8)
	bin.BigEndian.PutUint64(bs, x)
	return e.BytesTrim(bs)
}

func (e *hexEncoder) Address(address crypto.Address) string {
	return e.BytesTrim(address.Bytes())
}
