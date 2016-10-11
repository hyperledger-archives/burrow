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

package keys

import (
	"encoding/hex"
	"fmt"
)

type KeyClient interface {
	// Sign needs to return the signature bytes for given message to sign
	// and the address to sign it with.
	Sign(signBytesString string, signAddress []byte) (signature []byte, err error)
	// PublicKey needs to return the public key associated with a given address
	PublicKey(address []byte) (publicKey []byte, err error)
}

// NOTE [ben] Compiler check to ensure ErisKeyClient successfully implements
// eris-db/keys.KeyClient
var _ KeyClient = (*ErisKeyClient)(nil)

type ErisKeyClient struct {
	rpcString string
}

// ErisKeyClient.New returns a new eris-keys client for provided rpc location
// Eris-keys connects over http request-responses
func NewErisKeyClient(rpcString string) *ErisKeyClient {
	return &ErisKeyClient{
		rpcString: rpcString,
	}
}

// Eris-keys client Sign requests the signature from ErisKeysClient over rpc for the given
// bytes to be signed and the address to sign them with.
func (erisKeys *ErisKeyClient) Sign(signBytesString string, signAddress []byte) (signature []byte, err error) {
	args := map[string]string{
		"msg":  signBytesString,
		"hash": signBytesString, // TODO:[ben] backwards compatibility
		"addr": fmt.Sprintf("%X", signAddress),
	}
	sigS, err := RequestResponse(erisKeys.rpcString, "sign", args)
	if err != nil {
		return
	}
	sigBytes, err := hex.DecodeString(sigS)
	if err != nil {
		return
	}
	return sigBytes, err
}

// Eris-keys client PublicKey requests the public key associated with an address from
// the eris-keys server.
func (erisKeys *ErisKeyClient) PublicKey(address []byte) (publicKey []byte, err error) {
	args := map[string]string{
		"addr": fmt.Sprintf("%X", address),
	}
	pubS, err := RequestResponse(erisKeys.rpcString, "pub", args)
	if err != nil {
		return
	}
	// TODO: [ben] assert that received public key results in
	// address
	return hex.DecodeString(pubS)
}
