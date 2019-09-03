package encoding

import (
	"encoding/hex"
	fmt "fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/burrow/crypto"
)

const (
	HexPrefix = `0x`
)

func HexEncodeBytes(data []byte) string {
	return HexAddPrefix(hex.EncodeToString(data))
}

func HexEncodeNumber(i uint64) string {
	return HexAddPrefix(strconv.FormatUint(i, 16))
}

func HexAddPrefix(input string) string {
	return fmt.Sprintf("%s%s", HexPrefix, input)
}

func HexRemovePrefix(input string) string {
	return strings.Replace(input, HexPrefix, "", -1)
}

func HexDecodeToBytes(input string) ([]byte, error) {
	input = HexRemovePrefix(input)
	return hex.DecodeString(input)
}

func HexDecodeToNumber(i string) (uint64, error) {
	return strconv.ParseUint(i, 0, 64)
}

func HexDecodeToAddress(input string) (crypto.Address, error) {
	input = HexRemovePrefix(input)
	return crypto.AddressFromHexString(input)
}
