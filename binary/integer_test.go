package binary

import (
	"math"
	"math/big"
	"testing"

	"strconv"
	"strings"

	"fmt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsUint64SumOverflow(t *testing.T) {
	var b uint64 = 0xdeadbeef
	var a uint64 = math.MaxUint64 - b
	assert.False(t, IsUint64SumOverflow(a-b, b))
	assert.False(t, IsUint64SumOverflow(a, b))
	assert.False(t, IsUint64SumOverflow(a+b, 0))
	assert.True(t, IsUint64SumOverflow(a, b+1))
	assert.True(t, IsUint64SumOverflow(a+b, 1))
	assert.True(t, IsUint64SumOverflow(a+1, b+1))
}

func zero() *big.Int {
	return new(big.Int)
}

var big2E255 = zero().Lsh(big1, 255)
var big2E256 = zero().Lsh(big1, 256)
var big2E257 = zero().Lsh(big1, 257)

func TestU256(t *testing.T) {
	expected := big2E255
	encoded := U256(expected)
	assertBigIntEqual(t, expected, encoded, "Top bit set big int is fixed point")

	expected = zero()
	encoded = U256(big2E256)
	assertBigIntEqual(t, expected, encoded, "Ceiling bit is exact overflow")

	expected = zero().Sub(big2E256, big1)
	encoded = U256(expected)
	assertBigIntEqual(t, expected, encoded, "Max unsigned big int is fixed point")

	expected = big1
	encoded = U256(zero().Add(big2E256, big1))
	assertBigIntEqual(t, expected, encoded, "Overflow by one")

	expected = big2E255
	encoded = U256(zero().Add(big2E256, big2E255))
	assertBigIntEqual(t, expected, encoded, "Overflow by doubling")

	negative := big.NewInt(-234)
	assert.Equal(t, "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16",
		fmt.Sprintf("%X", U256(negative).Bytes()), "byte representation is twos complement")

	expected, ok := zero().SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16", 16)
	require.True(t, ok)
	assertBigIntEqual(t, expected, U256(negative), "Bytes representation should be twos complement")

	expected = zero()
	encoded = U256(zero().Neg(big2E256))
	assertBigIntEqual(t, expected, encoded, "Floor bit is overflow")

	expected = big2E255
	encoded = zero().Neg(big2E255)
	encoded = U256(encoded)
	assertBigIntEqual(t, expected, encoded, "2**255 is Self complement")

	expected = zero().Add(big2E255, big1)
	encoded = zero().Neg(big2E255)
	encoded = encoded.Add(encoded, big1)
	encoded = U256(encoded)
	assertBigIntEqual(t, expected, encoded, "")
}

func TestS256(t *testing.T) {
	expected := zero().Neg(big2E255)
	signed := S256(big2E255)
	assertBigIntEqual(t, expected, signed, "Should be negative")

	expected = zero().Sub(big2E255, big1)
	signed = S256(expected)
	assertBigIntEqual(t, expected, signed, "Maximum twos complement positive is fixed point")

	expected = zero()
	signed = S256(expected)
	assertBigIntEqual(t, expected, signed, "Twos complement of zero is fixed poount")

	// Technically undefined but let's not let that stop us
	expected = zero().Sub(big2E257, big2E256)
	signed = S256(big2E257)
	assertBigIntEqual(t, expected, signed, "Out of twos complement bounds")
}

func TestSignExtend(t *testing.T) {
	assertSignExtend(t, 16, 0,
		"0000 0000 1001 0000",
		"1111 1111 1001 0000")

	assertSignExtend(t, 16, 1,
		"1001 0000",
		"1001 0000")

	assertSignExtend(t, 32, 2,
		"0000 0000 1000 0000 1101 0011 1001 0000",
		"1111 1111 1000 0000 1101 0011 1001 0000")

	assertSignExtend(t, 32, 2,
		"0000 0000 0000 0000 1101 0011 1001 0000",
		"0000 0000 0000 0000 1101 0011 1001 0000")

	// Here we have a stray bit set in the 4th most significant byte that gets wiped out
	assertSignExtend(t, 32, 2,
		"0001 0000 0000 0000 1101 0011 1001 0000",
		"0000 0000 0000 0000 1101 0011 1001 0000")
	assertSignExtend(t, 32, 2,
		"0001 0000 1000 0000 1101 0011 1001 0000",
		"1111 1111 1000 0000 1101 0011 1001 0000")

	assertSignExtend(t, 32, 3,
		"0001 0000 1000 0000 1101 0011 1001 0000",
		"0001 0000 1000 0000 1101 0011 1001 0000")

	assertSignExtend(t, 32, 3,
		"1001 0000 1000 0000 1101 0011 1001 0000",
		"1001 0000 1000 0000 1101 0011 1001 0000")

	assertSignExtend(t, 64, 3,
		"0000 0000 0000 0000 0000 0000 0000 0000 1001 0000 1000 0000 1101 0011 1001 0000",
		"1111 1111 1111 1111 1111 1111 1111 1111 1001 0000 1000 0000 1101 0011 1001 0000")

	assertSignExtend(t, 64, 3,
		"0000 0000 0000 0000 0000 0000 0000 0000 0001 0000 1000 0000 1101 0011 1001 0000",
		"0000 0000 0000 0000 0000 0000 0000 0000 0001 0000 1000 0000 1101 0011 1001 0000")
}

func assertSignExtend(t *testing.T, bitSize int, bytesBack uint64, inputString, expectedString string) bool {
	input := intFromString(t, bitSize, inputString)
	expected := intFromString(t, bitSize, expectedString)
	//actual := SignExtend(big.NewInt(bytesBack), big.NewInt(int64(input)))
	actual := SignExtend(bytesBack, big.NewInt(int64(input)))
	var ret bool
	switch bitSize {
	case 8:
		ret = assert.Equal(t, uint8(expected), uint8(actual.Int64()))
	case 16:
		ret = assert.Equal(t, uint16(expected), uint16(actual.Int64()))
	case 32:
		ret = assert.Equal(t, uint32(expected), uint32(actual.Int64()))
	case 64:
		ret = assert.Equal(t, uint64(expected), uint64(actual.Int64()))
	default:
		t.Fatalf("Cannot test SignExtend for non-Go-native bit size %v", bitSize)
		return false
	}
	if !ret {

	}
	return ret
}

func assertBigIntEqual(t *testing.T, expected, actual *big.Int, messages ...string) bool {
	return assert.True(t, expected.Cmp(actual) == 0, fmt.Sprintf("%s - not equal:\n%v (expected)\n%v (actual)",
		strings.Join(messages, " "), expected, actual))
}

func intFromString(t *testing.T, bitSize int, binStrings ...string) uint64 {
	binaryString := strings.Replace(strings.Join(binStrings, ""), " ", "", -1)
	i, err := strconv.ParseUint(binaryString, 2, bitSize)
	require.NoError(t, err)
	return i
}
