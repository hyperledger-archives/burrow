package native

import (
	"crypto/sha256"

	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/permission"
	"golang.org/x/crypto/ripemd160"
)

var Precompiles = New().
	MustFunction(`Compute the sha256 hash of input`,
		leftPadAddress(2),
		permission.None,
		sha256Func).
	MustFunction(`Compute the ripemd160 hash of input`,
		leftPadAddress(3),
		permission.None,
		ripemd160Func).
	MustFunction(`Return an output identical to the input`,
		leftPadAddress(4),
		permission.None,
		identityFunc)

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

func sha256Func(ctx Context) (output []byte, err error) {
	// Deduct gas
	gasRequired := uint64((len(ctx.Input)+31)/32)*GasSha256Word + GasSha256Base
	if *ctx.Gas < gasRequired {
		return nil, errors.ErrorCodeInsufficientGas
	} else {
		*ctx.Gas -= gasRequired
	}
	// Hash
	hasher := sha256.New()
	// CONTRACT: this does not err
	hasher.Write(ctx.Input)
	return hasher.Sum(nil), nil
}

func ripemd160Func(ctx Context) (output []byte, err error) {
	// Deduct gas
	gasRequired := uint64((len(ctx.Input)+31)/32)*GasRipemd160Word + GasRipemd160Base
	if *ctx.Gas < gasRequired {
		return nil, errors.ErrorCodeInsufficientGas
	} else {
		*ctx.Gas -= gasRequired
	}
	// Hash
	hasher := ripemd160.New()
	// CONTRACT: this does not err
	hasher.Write(ctx.Input)
	return binary.LeftPadBytes(hasher.Sum(nil), 32), nil
}

func identityFunc(ctx Context) (output []byte, err error) {
	// Deduct gas
	gasRequired := uint64((len(ctx.Input)+31)/32)*GasIdentityWord + GasIdentityBase
	if *ctx.Gas < gasRequired {
		return nil, errors.ErrorCodeInsufficientGas
	} else {
		*ctx.Gas -= gasRequired
	}
	// Return identity
	return ctx.Input, nil
}

func leftPadAddress(bs ...byte) crypto.Address {
	return crypto.AddressFromWord256(binary.LeftPadWord256(bs))
}
