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

package keys

import (
	"encoding/hex"
	"fmt"

	"github.com/hyperledger/burrow/logging"
	logging_types "github.com/hyperledger/burrow/logging/types"
)

type KeyClient interface {
	// Sign needs to return the signature bytes for given message to sign
	// and the address to sign it with.
	Sign(signBytesString string, signAddress []byte) (signature []byte, err error)
	// PublicKey needs to return the public key associated with a given address
	PublicKey(address []byte) (publicKey []byte, err error)
}

// NOTE [ben] Compiler check to ensure monaxKeyClient successfully implements
// burrow/keys.KeyClient
var _ KeyClient = (*monaxKeyClient)(nil)

type monaxKeyClient struct {
	rpcString string
	logger    logging_types.InfoTraceLogger
}

// monaxKeyClient.New returns a new monax-keys client for provided rpc location
// Monax-keys connects over http request-responses
func NewBurrowKeyClient(rpcString string, logger logging_types.InfoTraceLogger) *monaxKeyClient {
	return &monaxKeyClient{
		rpcString: rpcString,
		logger:    logging.WithScope(logger, "BurrowKeyClient"),
	}
}

// Monax-keys client Sign requests the signature from BurrowKeysClient over rpc for the given
// bytes to be signed and the address to sign them with.
func (monaxKeys *monaxKeyClient) Sign(signBytesString string, signAddress []byte) (signature []byte, err error) {
	args := map[string]string{
		"msg":  signBytesString,
		"hash": signBytesString, // TODO:[ben] backwards compatibility
		"addr": fmt.Sprintf("%X", signAddress),
	}
	sigS, err := RequestResponse(monaxKeys.rpcString, "sign", args, monaxKeys.logger)
	if err != nil {
		return
	}
	sigBytes, err := hex.DecodeString(sigS)
	if err != nil {
		return
	}
	return sigBytes, err
}

// Monax-keys client PublicKey requests the public key associated with an address from
// the monax-keys server.
func (monaxKeys *monaxKeyClient) PublicKey(address []byte) (publicKey []byte, err error) {
	args := map[string]string{
		"addr": fmt.Sprintf("%X", address),
	}
	pubS, err := RequestResponse(monaxKeys.rpcString, "pub", args, monaxKeys.logger)
	if err != nil {
		return
	}
	// TODO: [ben] assert that received public key results in
	// address
	return hex.DecodeString(pubS)
}
