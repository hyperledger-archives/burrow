package wire

import (
	"bytes"
	"testing"
)

func TestReadStringEquality(t *testing.T) {

	var buf = bytes.NewBuffer(nil)
	var bufBytes []byte

	// Write a string
	var testString = "ThisIsSomeTestString"
	var n int
	var err error
	WriteString(testString, buf, &n, &err)
	if err != nil {
		t.Error(err.Error())
	}
	bufBytes = buf.Bytes()

	// Read the string, should return the same string
	buf = bytes.NewBuffer(bufBytes)
	var n2 int
	res := ReadString(buf, 0, &n2, &err)
	if err != nil {
		t.Error(err.Error())
	}
	if n != n2 {
		t.Error("Read string did not match write string length")
	}

	if testString != res {
		t.Error("Returned the wrong string")
	}

}

func TestReadStringLimit(t *testing.T) {

	var buf = bytes.NewBuffer(nil)
	var bufBytes []byte

	// Write a byteslice
	var testString = string("ThisIsSomeTestString")
	var n int
	var err error
	WriteString(testString, buf, &n, &err)
	if err != nil {
		t.Error(err.Error())
	}
	bufBytes = buf.Bytes()

	// Read the string, should work fine with no limit.
	buf = bytes.NewBuffer(bufBytes)
	var n2 int
	ReadString(buf, 0, &n2, &err)
	if err != nil {
		t.Error(err.Error())
	}
	if n != n2 {
		t.Error("Read string did not match write string length")
	}

	// Limit to the string length, should succeed.
	buf = bytes.NewBuffer(bufBytes)
	t.Logf("%X", bufBytes)
	var n3 int
	ReadString(buf, len(bufBytes), &n3, &err)
	if err != nil {
		t.Error(err.Error())
	}
	if n != n3 {
		t.Error("Read string did not match write string length")
	}

	// Limit to the string length, should succeed.
	buf = bytes.NewBuffer(bufBytes)
	var n4 int
	ReadString(buf, len(bufBytes)-1, &n4, &err)
	if err != ErrBinaryReadSizeOverflow {
		t.Error("Expected ErrBinaryReadsizeOverflow")
	}

}
