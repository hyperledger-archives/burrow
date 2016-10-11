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

package rpc_v0

import (
	"io"
	"io/ioutil"

	wire "github.com/tendermint/go-wire"

	rpc "github.com/eris-ltd/eris-db/rpc"
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
	var n int
	wire.WriteJSON(v, w, &n, &err)
	return err
}

// Encode to a byte array.
func (this *TCodec) EncodeBytes(v interface{}) ([]byte, error) {
	return wire.JSONBytes(v), nil
}

// Decode from an io.Reader.
func (this *TCodec) Decode(v interface{}, r io.Reader) error {
	bts, errR := ioutil.ReadAll(r)
	if errR != nil {
		return errR
	}
	var err error
	wire.ReadJSON(v, bts, &err)
	return err
}

// Decode from a byte array.
func (this *TCodec) DecodeBytes(v interface{}, bts []byte) error {
	var err error
	wire.ReadJSON(v, bts, &err)
	return err
}
