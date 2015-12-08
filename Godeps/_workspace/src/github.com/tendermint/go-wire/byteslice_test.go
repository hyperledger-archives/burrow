package wire

import (
	"bytes"
	"testing"
)

func TestReadByteSliceEquality(t *testing.T) {

	var buf = bytes.NewBuffer(nil)
	var bufBytes []byte

	// Write a byteslice
	var testBytes = []byte("ThisIsSomeTestArray")
	var n int
	var err error
	WriteByteSlice(testBytes, buf, &n, &err)
	if err != nil {
		t.Error(err.Error())
	}
	bufBytes = buf.Bytes()

	// Read the byteslice, should return the same byteslice
	buf = bytes.NewBuffer(bufBytes)
	var n2 int
	res := ReadByteSlice(buf, 0, &n2, &err)
	if err != nil {
		t.Error(err.Error())
	}
	if n != n2 {
		t.Error("Read bytes did not match write bytes length")
	}

	if !bytes.Equal(testBytes, res) {
		t.Error("Returned the wrong bytes")
	}

}

func TestReadByteSliceLimit(t *testing.T) {

	var buf = bytes.NewBuffer(nil)
	var bufBytes []byte

	// Write a byteslice
	var testBytes = []byte("ThisIsSomeTestArray")
	var n int
	var err error
	WriteByteSlice(testBytes, buf, &n, &err)
	if err != nil {
		t.Error(err.Error())
	}
	bufBytes = buf.Bytes()

	// Read the byteslice, should work fine with no limit.
	buf = bytes.NewBuffer(bufBytes)
	var n2 int
	ReadByteSlice(buf, 0, &n2, &err)
	if err != nil {
		t.Error(err.Error())
	}
	if n != n2 {
		t.Error("Read bytes did not match write bytes length")
	}

	// Limit to the byteslice length, should succeed.
	buf = bytes.NewBuffer(bufBytes)
	t.Logf("%X", bufBytes)
	var n3 int
	ReadByteSlice(buf, len(bufBytes), &n3, &err)
	if err != nil {
		t.Error(err.Error())
	}
	if n != n3 {
		t.Error("Read bytes did not match write bytes length")
	}

	// Limit to the byteslice length, should succeed.
	buf = bytes.NewBuffer(bufBytes)
	var n4 int
	ReadByteSlice(buf, len(bufBytes)-1, &n4, &err)
	if err != ErrBinaryReadSizeOverflow {
		t.Error("Expected ErrBinaryReadsizeOverflow")
	}

}
