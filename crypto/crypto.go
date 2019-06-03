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
)

type CurveType uint32

const (
	CurveTypeUnset CurveType = iota
	CurveTypeEd25519
	CurveTypeSecp256k1
)

func (k CurveType) String() string {
	switch k {
	case CurveTypeSecp256k1:
		return "secp256k1"
	case CurveTypeEd25519:
		return "ed25519"
	case CurveTypeUnset:
		return ""
	default:
		return "unknown"
	}
}
func (k CurveType) ABCIType() string {
	switch k {
	case CurveTypeSecp256k1:
		return "secp256k1"
	case CurveTypeEd25519:
		return "ed25519"
	case CurveTypeUnset:
		return ""
	default:
		return "unknown"
	}
}

// Get this CurveType's 8 bit identifier as a byte
func (k CurveType) Byte() byte {
	return byte(k)
}

func CurveTypeFromString(s string) (CurveType, error) {
	switch s {
	case "secp256k1":
		return CurveTypeSecp256k1, nil
	case "ed25519":
		return CurveTypeEd25519, nil
	case "":
		return CurveTypeUnset, nil
	default:
		return CurveTypeUnset, ErrInvalidCurve(s)
	}
}

type ErrInvalidCurve string

func (err ErrInvalidCurve) Error() string {
	return fmt.Sprintf("invalid curve type")
}

// The types in this file allow us to control serialisation of keys and signatures, as well as the interface
// exposed regardless of crypto library

type Signer interface {
	Sign(msg []byte) (*Signature, error)
}

// Signable is an interface for all signable things.
// It typically removes signatures before serializing.
type Signable interface {
	SignBytes(chainID string) ([]byte, error)
}
