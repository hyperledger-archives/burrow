package native

import (
	"crypto/sha256"
	"fmt"
	"math/big"

	"github.com/clearmatics/bn256"
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
		expModFunc).
	MustFunction(`Return the add of two points on a bn256 curve`,
		leftPadAddress(6),
		permission.None,
		bn256Add).
	MustFunction(`Return the scalar multiplication of a big int and a point on a bn256 curve`,
		leftPadAddress(7),
		permission.None,
		bn256ScalarMul).
	MustFunction(`Check the pairing of a set of points on a bn256 curve `,
		leftPadAddress(8),
		permission.None,
		bn256Pairing)

func leftPadAddress(bs ...byte) crypto.Address {
	return crypto.AddressFromWord256(binary.LeftPadWord256(bs))
}

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
	gasRequired := wordsIn(uint64(len(ctx.Input)))*GasSha256Word + GasSha256Base
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

func identityFunc(ctx Context) (output []byte, err error) {
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
func expModFunc(ctx Context) (output []byte, err error) {
	const errHeader = "expModFunc"

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

	base := new(big.Int).SetBytes(segments[0])
	exp := new(big.Int).SetBytes(segments[1])
	mod := new(big.Int).SetBytes(segments[2])

	// handle mod 0
	if mod.Sign() == 0 {
		return binary.LeftPadBytes([]byte{}, int(modLength)), nil
	}

	// return base**exp % mod left padded
	return binary.LeftPadBytes(new(big.Int).Exp(base, exp, mod).Bytes(), int(modLength)), nil
}

// bn256Add implements the EIP-196 for add pairs in a bn256 curve https://github.com/ethereum/EIPs/blob/master/EIPS/eip-196.md
func bn256Add(ctx Context) (output []byte, err error) {

	if *ctx.Gas < GasBn256Add {
		return nil, errors.Codes.InsufficientGas
	}
	*ctx.Gas -= GasBn256Add
	// retrieve the points from the input
	x := new(bn256.G1)
	y := new(bn256.G1)

	_, sgmnt, errs := cut(ctx.Input, binary.Word256Bytes*2, binary.Word256Bytes*2)
	if errs != nil {
		return nil, errs
	}

	_, errx := x.Unmarshal(sgmnt[0])
	if errx != nil {
		return nil, fmt.Errorf("error x: " + errx.Error())
	}

	_, erry := y.Unmarshal(sgmnt[1])
	if erry != nil {
		return nil, fmt.Errorf("error y: " + erry.Error())
	}
	//add them
	res := new(bn256.G1)
	res.Add(x, y)
	return res.Marshal(), nil
}

//bn256bn256ScalarMul implements the EIP-196 for scalar multiplication in a bn256 curve https://github.com/ethereum/EIPs/blob/master/EIPS/eip-196.md
func bn256ScalarMul(ctx Context) ([]byte, error) {
	if *ctx.Gas < GasBn256ScalarMul {
		return nil, errors.Codes.InsufficientGas
	}
	*ctx.Gas -= GasBn256ScalarMul

	//retrieve the point from the input
	_, sgmnt, errs := cut(ctx.Input, binary.Word256Bytes*2, binary.Word256Bytes)

	if errs != nil {
		return nil, errs
	}

	bnp := new(bn256.G1)
	_, errp := bnp.Unmarshal(sgmnt[0])
	if errp != nil {
		return nil, errp
	}
	//make the scalar multiplication
	res := new(bn256.G1)
	res.ScalarMult(bnp, new(big.Int).SetBytes(sgmnt[1]))
	return res.Marshal(), nil
}

// bn256Pairing implements the EIP-197 https://github.com/ethereum/EIPs/blob/master/EIPS/eip-197.md
func bn256Pairing(ctx Context) ([]byte, error) {
	if *ctx.Gas < GasBn256Pairing {
		return nil, errors.Codes.InsufficientGas
	}

	*ctx.Gas -= GasBn256Pairing

	// Handle some corner cases cheaply

	if len(ctx.Input)%192 > 0 {

		return nil, fmt.Errorf("bad elliptic curve pairing size")
	}
	// auxiliars for parse the inputs
	var (
		cs []*bn256.G1
		ts []*bn256.G2
	)
	// retrieving the inputs into the curve points
	for i := 0; i < len(ctx.Input); i += 192 {
		c, errc := newCurvePoint(ctx.Input[i : i+64])
		if errc != nil {
			return nil, errc
		}
		cs = append(cs, c)

		t, errt := newTwistPoint(ctx.Input[i+64 : i+192])
		if errt != nil {
			return nil, errt
		}
		ts = append(ts, t)
	}
	// check the parity
	return pairingCheckByte(cs, ts), nil
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

//represent bools as byte arrays of size 32. 1 for true, 0 for false
func pairingCheckByte(a []*bn256.G1, b []*bn256.G2) []byte {
	if bn256.PairingCheck(a, b) {
		return []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}

	}
	return make([]byte, binary.Word256Bytes)
}

func newCurvePoint(blob []byte) (*bn256.G1, error) {
	p := new(bn256.G1)
	if _, err := p.Unmarshal(blob); err != nil {
		return nil, err
	}
	return p, nil
}

// newTwistPoint unmarshals a binary blob into a bn256 elliptic curve point,
// returning it, or an error if the point is invalid.
func newTwistPoint(blob []byte) (*bn256.G2, error) {
	p := new(bn256.G2)
	if _, err := p.Unmarshal(blob); err != nil {
		return nil, err
	}
	return p, nil
}

func getUint64(bs []byte) uint64 {
	return binary.Uint64FromWord256(binary.LeftPadWord256(bs))
}

func wordsIn(numBytes uint64) uint64 {
	return numBytes + binary.Word256Bytes - 1/binary.Word256Bytes
}
