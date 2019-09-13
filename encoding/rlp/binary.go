package rlp

import (
	"encoding/binary"
)

func PutUint16(i uint64) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(i))
	return b
}

func PutUint24(i uint64) []byte {
	b := make([]byte, 3)
	b[0] = byte(i >> 16)
	b[1] = byte(i >> 8)
	b[2] = byte(i)
	return b
}

func PutUint32(i uint64) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(i))
	return b
}

func PutUint40(i uint64) []byte {
	b := make([]byte, 5)
	b[0] = byte(i >> 32)
	b[1] = byte(i >> 24)
	b[2] = byte(i >> 16)
	b[3] = byte(i >> 8)
	b[4] = byte(i)
	return b
}

func PutUint48(i uint64) []byte {
	b := make([]byte, 6)
	b[0] = byte(i >> 40)
	b[1] = byte(i >> 32)
	b[2] = byte(i >> 24)
	b[3] = byte(i >> 16)
	b[4] = byte(i >> 8)
	b[5] = byte(i)
	return b
}

func PutUint56(i uint64) []byte {
	b := make([]byte, 7)
	b[0] = byte(i >> 48)
	b[1] = byte(i >> 40)
	b[2] = byte(i >> 32)
	b[3] = byte(i >> 24)
	b[4] = byte(i >> 16)
	b[5] = byte(i >> 8)
	b[6] = byte(i)
	return b
}

func PutUint64(i uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(i))
	return b
}
