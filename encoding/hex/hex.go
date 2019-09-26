package encoding

import (
	"encoding/hex"
	fmt "fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/hyperledger/burrow/crypto"
)

const (
	HexPrefix = `0x`
)

func AddPrefix(input string) string {
	return fmt.Sprintf("%s%s", HexPrefix, input)
}

func RemovePrefix(input string) string {
	return strings.Replace(input, HexPrefix, "", -1)
}

func EncodeBytes(data []byte) string {
	return AddPrefix(hex.EncodeToString(data))
}

func EncodeNumber(i uint64) string {
	return AddPrefix(strconv.FormatUint(i, 16))
}

func DecodeToBytes(input string) ([]byte, error) {
	input = RemovePrefix(input)
	return hex.DecodeString(input)
}

func DecodeToNumber(i string) (uint64, error) {
	return strconv.ParseUint(i, 0, 64)
}

func DecodeToBigInt(input string) (*big.Int, error) {
	data, err := DecodeToBytes(input)
	if err != nil {
		return nil, err
	}
	return new(big.Int).SetBytes(data), nil
}

func DecodeToAddress(input string) (crypto.Address, error) {
	input = RemovePrefix(input)
	return crypto.AddressFromHexString(input)
}
