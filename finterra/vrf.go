// MIT License
//
// Copyright (c) 2018 Finterra
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//

package finterra

import (
	"encoding/hex"
	"math/big"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/execution/evm/sha3"
)

type VRF struct {
	acm.Addressable
	acm.Signer
	max256 *big.Int
	max    *big.Int
}

func NewVRF(addressable acm.Addressable, signer acm.Signer) VRF {
	vrf := VRF{
		Addressable: addressable,
		Signer:      signer,
		max:         big.NewInt(0),
		max256:      big.NewInt(0),
	}

	decMax, _ := hex.DecodeString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")

	vrf.max256.SetBytes(decMax)
	vrf.max.SetBytes(decMax)

	return vrf
}

func (vrf *VRF) SetMax(max uint64) {
	vrf.max.SetUint64(max)
}

// Evaluate returns a random number between 0 and 10^18 with the proof
func (vrf *VRF) Evaluate(m []byte) (index uint64, proof []byte) {
	// sign the hashed block height
	sig, err := vrf.Sign(m)

	if err != nil {
		return 0, nil
	}

	proof = make([]byte, 0)

	address := vrf.Address()
	proof = append(proof, address.Bytes()...)
	proof = append(proof, sig.Bytes()...)

	index = vrf.getIndex(sig.Bytes())

	return index, proof
}

// Verify ensure the proof is valid
func (vrf *VRF) Verify(m []byte, publicKey acm.PublicKey, proof []byte) (index uint64, result bool) {
	address, err := acm.AddressFromBytes(proof[0:binary.Word160Length])
	if err != nil {
		return 0, false
	}

	sig, err := acm.SignatureFromBytes(proof[binary.Word160Length+1:])
	if err != nil {
		return 0, false
	}

	// Verify address
	if publicKey.Address() != address {
		return 0, false
	}

	// Verify signature (proof)
	if !publicKey.VerifyBytes(m, sig) {
		return 0, false
	}

	index = vrf.getIndex(sig.Bytes())

	return index, true
}

func (vrf *VRF) getIndex(sig []byte) uint64 {
	hash := big.NewInt(0)
	hash.SetBytes(sha3.Sha3(sig))

	// construct the numerator and denominator for normalizing the signature uint between [0, 1]
	index := big.NewInt(0)
	numerator := big.NewInt(0)

	denominator := vrf.max256

	numerator = numerator.Mul(hash, vrf.max)

	// divide numerator and denominator to get the election ratio for this block height
	index = index.Div(numerator, denominator)

	return index.Uint64()
}
