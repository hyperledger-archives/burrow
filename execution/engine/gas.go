// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"math/big"

	"github.com/hyperledger/burrow/execution/errors"
)

const (
	GasSha3          uint64 = 1
	GasGetAccount    uint64 = 1
	GasStorageUpdate uint64 = 1
	GasCreateAccount uint64 = 1

	GasBaseOp  uint64 = 0 // TODO: make this 1
	GasStackOp uint64 = 1

	GasEcRecover     uint64 = 1
	GasSha256Word    uint64 = 1
	GasSha256Base    uint64 = 1
	GasRipemd160Word uint64 = 1
	GasRipemd160Base uint64 = 1
	GasExpModWord    uint64 = 1
	GasExpModBase    uint64 = 1
	GasIdentityWord  uint64 = 1
	GasIdentityBase  uint64 = 1
)

// Try to deduct gasToUse from gasLeft.  If ok return false, otherwise
// set err and return true.
func UseGasNegative(gasLeft *big.Int, gasToUse uint64) errors.CodedError {
	delta := new(big.Int).SetUint64(gasToUse)
	if gasLeft.Cmp(delta) >= 0 {
		gasLeft.Sub(gasLeft, delta)
	} else {
		return errors.Codes.InsufficientGas
	}
	return nil
}
