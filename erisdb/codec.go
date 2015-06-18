package erisdb

import (
	rpc "github.com/eris-ltd/eris-db/rpc"
	"github.com/tendermint/tendermint/binary"
	"io"
	"io/ioutil"
)

// Codec that uses tendermints 'binary' package for JSON.
type TCodec struct {
}

// Get a new codec.
func NewTCodec() rpc.Codec {
	return &TCodec{}
}

// Encode to an io.Writer.
func (this *TCodec) Encode(v interface{}, w io.Writer) error {
	var err error
	var n int64
	binary.WriteJSON(v, w, &n, &err)
	return err
}

// Encode to a byte array.
func (this *TCodec) EncodeBytes(v interface{}) ([]byte, error) {
	return binary.JSONBytes(v), nil
}

// Decode from an io.Reader.
func (this *TCodec) Decode(v interface{}, r io.Reader) error {
	bts, errR := ioutil.ReadAll(r)
	if errR != nil {
		return errR
	}
	var err error
	binary.ReadJSON(v, bts, &err)
	return err
}

// Decode from a byte array.
func (this *TCodec) DecodeBytes(v interface{}, bts []byte) error {
	var err error
	binary.ReadJSON(v, bts, &err)
	return err
}
