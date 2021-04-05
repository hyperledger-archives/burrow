package web3hex

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/hyperledger/burrow/crypto"
	"github.com/tmthrgd/go-hex"
)

type Decoder struct {
	error
	must bool
}

func (d *Decoder) Must() *Decoder {
	return &Decoder{must: true}
}

func (d *Decoder) Err() error {
	return d.error
}

func (d *Decoder) pushErr(err error) {
	if d.must {
		panic(err)
	}
	if d.error == nil {
		d.error = err
	}
}

func (d *Decoder) Bytes(hs string) []byte {
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

func (d *Decoder) Address(hs string) crypto.Address {
	if hs == "" {
		return crypto.Address{}
	}
	address, err := crypto.AddressFromBytes(d.Bytes(hs))
	if err != nil {
		d.pushErr(fmt.Errorf("could not decode address from '%s': %w", hs, err))
	}
	return address
}

func (d *Decoder) BigInt(hs string) *big.Int {
	return new(big.Int).SetBytes(d.Bytes(hs))
}

func (d *Decoder) Uint64(hs string) uint64 {
	bi := d.BigInt(hs)
	if !bi.IsUint64() {
		d.pushErr(fmt.Errorf("%v is not uint64", bi))
	}
	return bi.Uint64()
}

func (d *Decoder) Int64(hs string) int64 {
	bi := d.BigInt(hs)
	if !bi.IsInt64() {
		d.pushErr(fmt.Errorf("%v is not int64", bi))
	}
	return bi.Int64()
}
