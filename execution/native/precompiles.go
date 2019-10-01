package native

import (
	"crypto/sha256"
	"math/big"

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
		identityFunc).
	MustFunction(`Compute the operation base**exp % mod where the values are big ints`,
		leftPadAddress(5),
		permission.None,
		bigModExp)

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
	hashed := crypto.Keccak256(recovered[1:])
	return LeftPadBytes(hashed, 32), nil
}
*/

func sha256Func(ctx Context) (output []byte, err error) {
	// Deduct gas
	gasRequired := uint64((len(ctx.Input)+31)/32)*GasSha256Word + GasSha256Base
	if *ctx.Gas < gasRequired {
		return nil, errors.Codes.InsufficientGas
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
		return nil, errors.Codes.InsufficientGas
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
		return nil, errors.Codes.InsufficientGas
	} else {
		*ctx.Gas -= gasRequired
	}
	// Return identity
	return ctx.Input, nil
}

const (
	// gas requirement for bigModExp set to 1
	GasRequire uint64 = 1
)

// bigModExp: function that implement the EIP 198 (https://github.com/ethereum/EIPs/blob/master/EIPS/eip-198.md with a fixed gas requirement)
func bigModExp(ctx Context) (output []byte, err error) {

	if *ctx.Gas < GasRequire {
		return nil, errors.Codes.InsufficientGas
	}

	*ctx.Gas -= GasRequire
	// get the lengths of base, exp and mod
	baseLen := new(big.Int).SetBytes(binary.RightPadBytes(ctx.Input[0:32], 32)).Uint64()
	expLen := new(big.Int).SetBytes(binary.RightPadBytes(ctx.Input[32:64], 32)).Uint64()
	modLen := new(big.Int).SetBytes(binary.RightPadBytes(ctx.Input[64:96], 32)).Uint64()

	// shift input array to the actual values
	if len(ctx.Input) > 96 {
		ctx.Input = ctx.Input[96:]
	} else {
		ctx.Input = ctx.Input[:0]
	}

	// handle the case when tehre is no base nor mod
	if baseLen+modLen == 0 {
		return []byte{}, nil
	}

	// get the values of base, exp and mod
	base := new(big.Int).SetBytes(getData(ctx.Input, 0, baseLen))
	exp := new(big.Int).SetBytes(getData(ctx.Input, baseLen, expLen))
	mod := new(big.Int).SetBytes(getData(ctx.Input, baseLen+expLen, modLen))
	// handle mod 0
	if mod.Sign() == 0 {
		return binary.LeftPadBytes([]byte{}, int(modLen)), nil
	}
	// return base**exp % mod left padded
	return binary.LeftPadBytes(new(big.Int).Exp(base, exp, mod).Bytes(), int(modLen)), nil

}

// auxiliar function to retrieve data from arrays
func getData(data []byte, start uint64, size uint64) []byte {
	length := uint64(len(data))
	if start > length {
		start = length
	}
	end := start + size
	if end > length {
		end = length
	}
	return binary.RightPadBytes(data[start:end], int(size))
}

func leftPadAddress(bs ...byte) crypto.Address {
	return crypto.AddressFromWord256(binary.LeftPadWord256(bs))
}
