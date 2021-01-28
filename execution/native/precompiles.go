package native

import (
	cryptoSha256 "crypto/sha256"
	"fmt"
	"math/big"

	"github.com/btcsuite/btcd/btcec"

	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/permission"
	"golang.org/x/crypto/ripemd160"
)

var Precompiles = New().
	MustFunction(`Recover public key/address of account that signed the data`,
		leftPadAddress(1),
		permission.None,
		ecrecover).
	MustFunction(`Compute the sha256 hash of input`,
		leftPadAddress(2),
		permission.None,
		sha256).
	MustFunction(`Compute the ripemd160 hash of input`,
		leftPadAddress(3),
		permission.None,
		ripemd160Func).
	MustFunction(`Return an output identical to the input`,
		leftPadAddress(4),
		permission.None,
		identity).
	MustFunction(`Compute the operation base**exp % mod where the values are big ints`,
		leftPadAddress(5),
		permission.None,
		expMod)

func leftPadAddress(bs ...byte) crypto.Address {
	return crypto.AddressFromWord256(binary.LeftPadWord256(bs))
}

// SECP256K1 Recovery
func ecrecover(ctx Context) ([]byte, error) {
	// Deduct gas
	gasRequired := GasEcRecover
	if *ctx.Gas < gasRequired {
		return nil, errors.Codes.InsufficientGas
	} else {
		*ctx.Gas -= gasRequired
	}

	// layout is:
	// input:  [ hash |  v   |  r   |  s   ]
	// bytes:  [ 32   |  32  |  32  |  32  ]
	// Where:
	//   hash = message digest
	//   v = 27 + recovery id (which of 4 possible x coords do we take as public key) (single byte but padded)
	//   r = encrypted random point
	//   s = signature proof

	// Signature layout required by ethereum:
	// sig:    [  r   |  s   |  v  ]
	// bytes:  [  32  |  32  |  1  ]
	hash := ctx.Input[:32]

	const compactSigLength = 2*binary.Word256Bytes + 1
	sig := make([]byte, compactSigLength)
	// Copy in r, s
	copy(sig, ctx.Input[2*binary.Word256Bytes:4*binary.Word256Bytes])
	// Check v is single byte
	v := ctx.Input[binary.Word256Bytes : 2*binary.Word256Bytes]
	if !binary.IsZeros(v[:len(v)-1]) {
		return nil, fmt.Errorf("ecrecover: recovery ID is larger than one byte")
	}
	// Copy in v to last element of sig
	sig[2*binary.Word256Bytes] = v[len(v)-1]

	publicKey, isCompressed, err := btcec.RecoverCompact(btcec.S256(), sig, hash)
	if err != nil {
		return nil, err
	}

	var serializedPublicKey []byte
	if isCompressed {
		serializedPublicKey = publicKey.SerializeCompressed()
	} else {
		serializedPublicKey = publicKey.SerializeUncompressed()
	}
	// First byte is a length-prefix
	hashed := crypto.Keccak256(serializedPublicKey[1:])
	hashed = hashed[len(hashed)-crypto.AddressLength:]
	return binary.LeftPadBytes(hashed, binary.Word256Bytes), nil
}

func sha256(ctx Context) (output []byte, err error) {
	// Deduct gas
	gasRequired := wordsIn(uint64(len(ctx.Input)))*GasSha256Word + GasSha256Base
	if *ctx.Gas < gasRequired {
		return nil, errors.Codes.InsufficientGas
	} else {
		*ctx.Gas -= gasRequired
	}
	// Hash
	hasher := cryptoSha256.New()
	// CONTRACT: this does not err
	hasher.Write(ctx.Input)
	return hasher.Sum(nil), nil
}

func ripemd160Func(ctx Context) (output []byte, err error) {
	// Deduct gas
	gasRequired := wordsIn(uint64(len(ctx.Input)))*GasRipemd160Word + GasRipemd160Base
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

func identity(ctx Context) (output []byte, err error) {
	// Deduct gas
	gasRequired := wordsIn(uint64(len(ctx.Input)))*GasIdentityWord + GasIdentityBase
	if *ctx.Gas < gasRequired {
		return nil, errors.Codes.InsufficientGas
	} else {
		*ctx.Gas -= gasRequired
	}
	// Return identity
	return ctx.Input, nil
}

// expMod: function that implements the EIP 198 (https://github.com/ethereum/EIPs/blob/master/EIPS/eip-198.md with
// a fixed gas requirement)
func expMod(ctx Context) (output []byte, err error) {
	const errHeader = "expMod"

	input, segments, err := cut(ctx.Input, binary.Word256Bytes, binary.Word256Bytes, binary.Word256Bytes)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", errHeader, err)
	}

	// get the lengths of base, exp and mod
	baseLength := getUint64(segments[0])
	expLength := getUint64(segments[1])
	modLength := getUint64(segments[2])

	// TODO: implement non-trivial gas schedule for this operation. Probably a parameterised version of the one
	// described in EIP though that one seems like a bit of a complicated fudge
	gasRequired := GasExpModBase + GasExpModWord*(wordsIn(baseLength)*wordsIn(expLength)*wordsIn(modLength))

	if *ctx.Gas < gasRequired {
		return nil, errors.Codes.InsufficientGas
	}

	*ctx.Gas -= gasRequired

	input, segments, err = cut(input, baseLength, expLength, modLength)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", errHeader, err)
	}

	// get the values of base, exp and mod
	base := getBigInt(segments[0], baseLength)
	exp := getBigInt(segments[1], expLength)
	mod := getBigInt(segments[2], modLength)

	// handle mod 0
	if mod.Sign() == 0 {
		return binary.LeftPadBytes([]byte{}, int(modLength)), nil
	}

	// return base**exp % mod left padded
	return binary.LeftPadBytes(new(big.Int).Exp(base, exp, mod).Bytes(), int(modLength)), nil
}

// Partition the head of input into segments for each length in lengths. The first return value is the unconsumed tail
// of input and the seconds is the segments. Returns an error if input is of insufficient length to establish each segment.
func cut(input []byte, lengths ...uint64) ([]byte, [][]byte, error) {
	segments := make([][]byte, len(lengths))
	for i, length := range lengths {
		if uint64(len(input)) < length {
			return nil, nil, fmt.Errorf("input is not long enough")
		}
		segments[i] = input[:length]
		input = input[length:]
	}
	return input, segments, nil
}

func getBigInt(bs []byte, numBytes uint64) *big.Int {
	bits := uint(numBytes) * 8
	// Push bytes into big.Int and interpret as twos complement encoding with of bits width
	return binary.FromTwosComplement(new(big.Int).SetBytes(bs), bits)
}

func getUint64(bs []byte) uint64 {
	return binary.Uint64FromWord256(binary.LeftPadWord256(bs))
}

func wordsIn(numBytes uint64) uint64 {
	return numBytes + binary.Word256Bytes - 1/binary.Word256Bytes
}
