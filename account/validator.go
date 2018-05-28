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

package account

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
)

type Validator interface {
	Addressable

	Power() int64
	Stake() uint64
	Sequence() uint64
	BondingHeight() uint64

	MinimumStakeToUnbond() uint64

	Bytes() ([]byte, error)
	String() string
}

type MutableValidator interface {
	Validator

	AddStake(stake uint64)
	SubtractStake(stake uint64)
	IncSequence()
}

type validator struct {
	publicKey     PublicKey
	stake         uint64
	bondingHeight uint64
	sequence      uint64
}

func NewValidator(publicKey PublicKey, stake, bondingHeight uint64) Validator {
	return validator{
		publicKey:     publicKey,
		stake:         stake,
		bondingHeight: bondingHeight,
		sequence:      0,
	}
}

func LoadValidator(bytes []byte) Validator {
	u := map[string]string{}
	err := json.Unmarshal(bytes, &u)
	if err != nil {
		return nil
	}

	publicKeyBytes, err := hex.DecodeString(u["publicKey"])
	if err != nil {
		return nil
	}

	publicKey, err := PublicKeyFromBytes(publicKeyBytes[1:])
	if err != nil {
		return nil
	}

	stake, err := strconv.ParseUint(u["stake"], 10, 64)
	if err != nil {
		return nil
	}

	bondingHeight, err := strconv.ParseUint(u["bondingHeight"], 10, 64)
	if err != nil {
		return nil
	}

	sequence, err := strconv.ParseUint(u["sequence"], 10, 64)
	if err != nil {
		return nil
	}

	return validator{
		publicKey:     publicKey,
		stake:         stake,
		bondingHeight: bondingHeight,
		sequence:      sequence,
	}
}

func (val validator) Bytes() ([]byte, error) {
	u := map[string]string{}
	u["publicKey"] = hex.EncodeToString(val.publicKey.Bytes())
	u["stake"] = strconv.FormatUint(val.stake, 10)
	u["bondingHeight"] = strconv.FormatUint(val.bondingHeight, 10)
	u["sequence"] = strconv.FormatUint(val.sequence, 10)

	bytes, err := json.Marshal(u)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (val validator) String() string {
	return fmt.Sprintf("{Address:%v, Stake:%v, BondingHeight:%v}", val.Address(), val.Stake(), val.BondingHeight())
}

func (val validator) Address() Address {
	return val.publicKey.Address()
}

func (val validator) PublicKey() PublicKey {
	return val.publicKey
}

func (val validator) Power() int64 {
	// Viva democracy, every person will be treated equally in our blockchain
	return 1
}

func (val validator) Stake() uint64 {
	return val.stake
}

func (val validator) BondingHeight() uint64 {
	return val.bondingHeight
}

func (val validator) Sequence() uint64 {
	return val.sequence
}

func (val validator) MinimumStakeToUnbond() uint64 {
	//TODO:Mostafa
	return 0
}

func (val *validator) AddStake(stake uint64) {
	val.stake += stake
}

func (val *validator) SubtractStake(stake uint64) {
	if val.stake < stake {
		val.stake = 0
	} else {
		val.stake -= stake
	}
}

func (val *validator) IncSequence() {
	val.sequence++
}

func AsMutableValidator(val Validator) MutableValidator {
	if val == nil {
		return nil
	}

	return &validator{
		publicKey:     val.PublicKey(),
		stake:         val.Stake(),
		bondingHeight: val.BondingHeight(),
		sequence:      val.Sequence(),
	}
}
