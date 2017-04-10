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

	logging_types "github.com/monax/burrow/logging/types"
	"github.com/monax/burrow/logging"
)

type KeyClient interface {
	// Sign needs to return the signature bytes for given message to sign
	// and the address to sign it with.
	Sign(signBytesString string, signAddress []byte) (signature []byte, err error)
	// PublicKey needs to return the public key associated with a given address
	PublicKey(address []byte) (publicKey []byte, err error)
}

// NOTE [ben] Compiler check to ensure erisKeyClient successfully implements
// eris-db/keys.KeyClient
var _ KeyClient = (*erisKeyClient)(nil)

type erisKeyClient struct {
	rpcString string
	logger    logging_types.InfoTraceLogger
}

// erisKeyClient.New returns a new eris-keys client for provided rpc location
// Eris-keys connects over http request-responses
func NewErisKeyClient(rpcString string, logger logging_types.InfoTraceLogger) *erisKeyClient {
	return &erisKeyClient{
		rpcString: rpcString,
		logger:    logging.WithScope(logger, "ErisKeysClient"),
	}
}

// Eris-keys client Sign requests the signature from ErisKeysClient over rpc for the given
// bytes to be signed and the address to sign them with.
func (erisKeys *erisKeyClient) Sign(signBytesString string, signAddress []byte) (signature []byte, err error) {
	args := map[string]string{
		"msg":  signBytesString,
		"hash": signBytesString, // TODO:[ben] backwards compatibility
		"addr": fmt.Sprintf("%X", signAddress),
	}
	sigS, err := RequestResponse(erisKeys.rpcString, "sign", args, erisKeys.logger)
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
func (erisKeys *erisKeyClient) PublicKey(address []byte) (publicKey []byte, err error) {
	args := map[string]string{
		"addr": fmt.Sprintf("%X", address),
	}
	pubS, err := RequestResponse(erisKeys.rpcString, "pub", args, erisKeys.logger)
	if err != nil {
		return
	}
	// TODO: [ben] assert that received public key results in
	// address
	return hex.DecodeString(pubS)
}
