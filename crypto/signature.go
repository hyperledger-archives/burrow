// Copyright 2019 Monax Industries Limited
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

package crypto

import (
	"fmt"

	hex "github.com/tmthrgd/go-hex"
	"golang.org/x/crypto/ed25519"
)

func SignatureFromBytes(bs []byte, curveType CurveType) (*Signature, error) {
	switch curveType {
	case CurveTypeEd25519:
		if len(bs) != ed25519.SignatureSize {
			return nil, fmt.Errorf("bytes passed have length %v by ed25519 signatures have %v bytes",
				len(bs), ed25519.SignatureSize)
		}
	case CurveTypeSecp256k1:
		// TODO: validate?
	}

	return &Signature{CurveType: curveType, Signature: bs}, nil
}

func (sig *Signature) RawBytes() []byte {
	return sig.Signature
}

func (sig *Signature) String() string {
	return hex.EncodeUpperToString(sig.Signature)
}
