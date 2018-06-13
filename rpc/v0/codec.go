// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v0

import (
	"io"
	"io/ioutil"

	"encoding/json"

	"github.com/hyperledger/burrow/rpc"
)

// Codec that uses tendermints 'binary' package for JSON.
type TCodec struct {
}

// Get a new codec.
func NewTCodec() rpc.Codec {
	return &TCodec{}
}

// Encode to an io.Writer.
func (codec *TCodec) Encode(v interface{}, w io.Writer) error {
	bs, err := codec.EncodeBytes(v)
	if err != nil {
		return err
	}
	_, err = w.Write(bs)
	return err
}

// Encode to a byte array.
func (codec *TCodec) EncodeBytes(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Decode from an io.Reader.
func (codec *TCodec) Decode(v interface{}, r io.Reader) error {
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return codec.DecodeBytes(v, bs)
}

// Decode from a byte array.
func (codec *TCodec) DecodeBytes(v interface{}, bs []byte) error {
	return json.Unmarshal(bs, v)
}
