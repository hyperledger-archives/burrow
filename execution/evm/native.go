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

package evm

import (
	"crypto/sha256"
	"math"
	"math/big"

	. "github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/logging"
	"golang.org/x/crypto/ripemd160"
)

var registeredNativeContracts = make(map[crypto.Address]NativeContract)

func IsRegisteredNativeContract(address crypto.Address) bool {
	_, ok := registeredNativeContracts[address]
	return ok
}

func RegisterNativeContract(address crypto.Address, fn NativeContract) bool {
	_, exists := registeredNativeContracts[address]
	if exists {
		return false
	}
	registeredNativeContracts[address] = fn
	return true
}

func init() {
	registerNativeContracts()
	registerSNativeContracts()
}

var (
	natAddr1, _ = crypto.AddressFromBytes(LeftPadBytes([]byte{1}, 20))
	natAddr2, _ = crypto.AddressFromBytes(LeftPadBytes([]byte{2}, 20))
	natAddr3, _ = crypto.AddressFromBytes(LeftPadBytes([]byte{3}, 20))
	natAddr4, _ = crypto.AddressFromBytes(LeftPadBytes([]byte{4}, 20))
	natAddr5, _ = crypto.AddressFromBytes(LeftPadBytes([]byte{5}, 20))
)

func registerNativeContracts() {
	// registeredNativeContracts[Int64ToWord256(1)] = ecrecoverFunc
	registeredNativeContracts[natAddr2] = sha256Func
	registeredNativeContracts[natAddr3] = ripemd160Func
	registeredNativeContracts[natAddr4] = identityFunc
	registeredNativeContracts[natAddr5] = bigModExp
}

//-----------------------------------------------------------------------------

func ExecuteNativeContract(address crypto.Address, st Interface, caller crypto.Address, input []byte, gas *uint64,
	logger *logging.Logger) ([]byte, errors.CodedError) {

	contract, ok := registeredNativeContracts[address]
	if !ok {
		return nil, errors.ErrorCodef(errors.ErrorCodeNativeFunction,
			"no native contract registered at address: %v", address)
	}
	output, err := contract(st, caller, input, gas, logger)
	if err != nil {
		return nil, errors.NewException(errors.ErrorCodeNativeFunction, err.Error())
	}
	return output, nil
}

type NativeContract func(state Interface, caller crypto.Address, input []byte, gas *uint64,
	logger *logging.Logger) (output []byte, err error)

/* Removed due to C dependency
func ecrecoverFunc(state State, caller crypto.Address, input []byte, gas *int64) (output []byte, err error) {
	// Deduct gas
	gasRequired := GasEcRecover
	if *gas < gasRequired {
		return nil, ErrInsufficientGas
	} else {
		*gas -= gasRequired
	}
	// Recover
	hash := input[:32]
	v := byte(input[32] - 27) // ignore input[33:64], v is small.
	sig := append(input[64:], v)

	recovered, err := secp256k1.RecoverPubkey(hash, sig)
	if err != nil {
		return nil, err
OH NO STOCASTIC CAT CODING!!!!
	}
	hashed := sha3.Sha3(recovered[1:])
	return LeftPadBytes(hashed, 32), nil
}
*/

func sha256Func(state Interface, caller crypto.Address, input []byte, gas *uint64,
	logger *logging.Logger) (output []byte, err error) {
	// Deduct gas
	gasRequired := uint64((len(input)+31)/32)*GasSha256Word + GasSha256Base
	if *gas < gasRequired {
		return nil, errors.ErrorCodeInsufficientGas
	} else {
		*gas -= gasRequired
	}
	// Hash
	hasher := sha256.New()
	// CONTRACT: this does not err
	hasher.Write(input)
	return hasher.Sum(nil), nil
}

func ripemd160Func(state Interface, caller crypto.Address, input []byte, gas *uint64,
	logger *logging.Logger) (output []byte, err error) {
	// Deduct gas
	gasRequired := uint64((len(input)+31)/32)*GasRipemd160Word + GasRipemd160Base
	if *gas < gasRequired {
		return nil, errors.ErrorCodeInsufficientGas
	} else {
		*gas -= gasRequired
	}
	// Hash
	hasher := ripemd160.New()
	// CONTRACT: this does not err
	hasher.Write(input)
	return LeftPadBytes(hasher.Sum(nil), 32), nil
}

func identityFunc(state Interface, caller crypto.Address, input []byte, gas *uint64,
	logger *logging.Logger) (output []byte, err error) {
	// Deduct gas
	gasRequired := uint64((len(input)+31)/32)*GasIdentityWord + GasIdentityBase
	if *gas < gasRequired {
		return nil, errors.ErrorCodeInsufficientGas
	} else {
		*gas -= gasRequired
	}
	// Return identity
	return input, nil
}

// from go-ethereum/blob/master/core/vm/contracts.go
var (
	big1      = big.NewInt(1)
	big4      = big.NewInt(4)
	big8      = big.NewInt(8)
	big16     = big.NewInt(16)
	big32     = big.NewInt(32)
	big64     = big.NewInt(64)
	big96     = big.NewInt(96)
	big480    = big.NewInt(480)
	big1024   = big.NewInt(1024)
	big3072   = big.NewInt(3072)
	big199680 = big.NewInt(199680)
)

const (
	MaxUint64 = 1<<64 - 1

	ModExpQuadCoeffDiv uint64 = 20 // Divisor for the quadratic particle of the big int modular exponentiation
)

func bigModExp(state Interface, caller crypto.Address, input []byte, gas *uint64, logger *logging.Logger) (output []byte, err error) {
	//Deduct gas
	gasRequired := bigModExpGasRequire(input)
	if *gas < gasRequired {
		return nil, errors.ErrorCodeInsufficientGas
	}
	*gas -= gasRequired

	var (
		baseLen = new(big.Int).SetBytes(getData(input, 0, 32)).Uint64()
		expLen  = new(big.Int).SetBytes(getData(input, 32, 32)).Uint64()
		modLen  = new(big.Int).SetBytes(getData(input, 64, 32)).Uint64()
	)
	if len(input) > 96 {
		input = input[96:]
	} else {
		input = input[:0]
	}
	// Handle a special case when both the base and mod length is zero
	if baseLen == 0 && modLen == 0 {
		return []byte{}, nil
	}
	// Retrieve the operands and execute the exponentiation
	var (
		base = new(big.Int).SetBytes(getData(input, 0, baseLen))
		exp  = new(big.Int).SetBytes(getData(input, baseLen, expLen))
		mod  = new(big.Int).SetBytes(getData(input, baseLen+expLen, modLen))
	)
	if mod.BitLen() == 0 {
		// Modulo 0 is undefined, return zero
		return LeftPadBytes([]byte{}, int(modLen)), nil
	}
	return LeftPadBytes(base.Exp(base, exp, mod).Bytes(), int(modLen)), nil

}

// RequiredGas returns the gas required to execute the pre-compiled contract.
func bigModExpGasRequire(input []byte) uint64 {
	var (
		baseLen = new(big.Int).SetBytes(getData(input, 0, 32))
		expLen  = new(big.Int).SetBytes(getData(input, 32, 32))
		modLen  = new(big.Int).SetBytes(getData(input, 64, 32))
	)
	if len(input) > 96 {
		input = input[96:]
	} else {
		input = input[:0]
	}
	// Retrieve the head 32 bytes of exp for the adjusted exponent length
	var expHead *big.Int
	if big.NewInt(int64(len(input))).Cmp(baseLen) <= 0 {
		expHead = new(big.Int)
	} else {
		if expLen.Cmp(big32) > 0 {
			expHead = new(big.Int).SetBytes(getData(input, baseLen.Uint64(), 32))
		} else {
			expHead = new(big.Int).SetBytes(getData(input, baseLen.Uint64(), expLen.Uint64()))
		}
	}
	// Calculate the adjusted exponent length
	var msb int
	if bitlen := expHead.BitLen(); bitlen > 0 {
		msb = bitlen - 1
	}
	adjExpLen := new(big.Int)
	if expLen.Cmp(big32) > 0 {
		adjExpLen.Sub(expLen, big32)
		adjExpLen.Mul(big8, adjExpLen)
	}
	adjExpLen.Add(adjExpLen, big.NewInt(int64(msb)))

	// Calculate the gas cost of the operation
	gas := new(big.Int).Set(bigMax(modLen, baseLen))
	switch {
	case gas.Cmp(big64) <= 0:
		gas.Mul(gas, gas)
	case gas.Cmp(big1024) <= 0:
		gas = new(big.Int).Add(
			new(big.Int).Div(new(big.Int).Mul(gas, gas), big4),
			new(big.Int).Sub(new(big.Int).Mul(big96, gas), big3072),
		)
	default:
		gas = new(big.Int).Add(
			new(big.Int).Div(new(big.Int).Mul(gas, gas), big16),
			new(big.Int).Sub(new(big.Int).Mul(big480, gas), big199680),
		)
	}
	gas.Mul(gas, bigMax(adjExpLen, big1))
	gas.Div(gas, new(big.Int).SetUint64(ModExpQuadCoeffDiv))

	if gas.BitLen() > 64 {
		return math.MaxUint64
	}
	return gas.Uint64()
}

func getData(data []byte, start uint64, size uint64) []byte {
	length := uint64(len(data))
	if start > length {
		start = length
	}
	end := start + size
	if end > length {
		end = length
	}
	return RightPadBytes(data[start:end], int(size))
}

func bigMax(x, y *big.Int) *big.Int {
	if x.Cmp(y) < 0 {
		return y
	}
	return x
}
