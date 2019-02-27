package service

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/hyperledger/burrow/vent/logger"
	"github.com/stretchr/testify/assert"
	hex "github.com/tmthrgd/go-hex"
)

func TestUTF8StringFromBytes(t *testing.T) {
	// The code point for ó is less than 255 but needs two unicode bytes - it's value expressed as a single byte
	// is in the private use area so is invalid.
	goodString := "Cliente - Doc. identificación"
	badString := BadStringToHexFunction(goodString)
	str, err := UTF8StringFromBytes([]byte(badString))
	assert.Equal(t, "Cliente - Doc. identificaci�n", str)
	assert.Contains(t, err.Error(), "0xF3 (at index 27)")

	goodString += goodString
	badString = BadStringToHexFunction(goodString)
	str, err = UTF8StringFromBytes([]byte(badString))
	assert.Equal(t, "Cliente - Doc. identificaci�nCliente - Doc. identificaci�n", str)
	assert.Contains(t, err.Error(), "0xF3 (at index 27)")
	assert.Contains(t, err.Error(), "0xF3 (at index 56)")
}

func TestSanitiseBytesForString(t *testing.T) {
	goodString := "Cliente - Doc. identificación"
	badString := BadStringToHexFunction(goodString)
	str := sanitiseBytesForString([]byte(badString), logger.NewLogger("error"))
	assert.Equal(t, "Cliente - Doc. identificaci�n", str)
}

// Shared by consumer_test
func BadStringToHexFunction(goodString string) string {
	// real life example from an asciiToHex function intended to generate hex of a utf8 string
	buf := new(bytes.Buffer)
	for _, r := range goodString {
		// This is effectively the algorithm used by asciiToHex from burrow.js - this is broken!
		// will always create incorrect bytes for multi-byte utf8 code points and sometimes invalid utf8
		buf.WriteString(fmt.Sprintf("%2X", r))
	}
	return string(hex.MustDecodeString(buf.String()))
}
