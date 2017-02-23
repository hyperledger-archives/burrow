// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package random

import (
	crand "crypto/rand"
	"math/rand"
	"time"

	"github.com/eris-ltd/eris-db/common/sanity"
)

const (
	strChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" // 62 characters
)

func init() {
	b := cRandBytes(8)
	var seed uint64
	for i := 0; i < 8; i++ {
		seed |= uint64(b[i])
		seed <<= 8
	}
	rand.Seed(int64(seed))
}

// Constructs an alphanumeric string of given length.
func RandStr(length int) string {
	chars := []byte{}
MAIN_LOOP:
	for {
		val := rand.Int63()
		for i := 0; i < 10; i++ {
			v := int(val & 0x3f) // rightmost 6 bits
			if v >= 62 {         // only 62 characters in strChars
				val >>= 6
				continue
			} else {
				chars = append(chars, strChars[v])
				if len(chars) == length {
					break MAIN_LOOP
				}
				val >>= 6
			}
		}
	}

	return string(chars)
}

func RandUint16() uint16 {
	return uint16(rand.Uint32() & (1<<16 - 1))
}

func RandUint32() uint32 {
	return rand.Uint32()
}

func RandUint64() uint64 {
	return uint64(rand.Uint32())<<32 + uint64(rand.Uint32())
}

func RandUint() uint {
	return uint(rand.Int())
}

func RandInt16() int16 {
	return int16(rand.Uint32() & (1<<16 - 1))
}

func RandInt32() int32 {
	return int32(rand.Uint32())
}

func RandInt64() int64 {
	return int64(rand.Uint32())<<32 + int64(rand.Uint32())
}

func RandInt() int {
	return rand.Int()
}

// Distributed pseudo-exponentially to test for various cases
func RandUint16Exp() uint16 {
	bits := rand.Uint32() % 16
	if bits == 0 {
		return 0
	}
	n := uint16(1 << (bits - 1))
	n += uint16(rand.Int31()) & ((1 << (bits - 1)) - 1)
	return n
}

// Distributed pseudo-exponentially to test for various cases
func RandUint32Exp() uint32 {
	bits := rand.Uint32() % 32
	if bits == 0 {
		return 0
	}
	n := uint32(1 << (bits - 1))
	n += uint32(rand.Int31()) & ((1 << (bits - 1)) - 1)
	return n
}

// Distributed pseudo-exponentially to test for various cases
func RandUint64Exp() uint64 {
	bits := rand.Uint32() % 64
	if bits == 0 {
		return 0
	}
	n := uint64(1 << (bits - 1))
	n += uint64(rand.Int63()) & ((1 << (bits - 1)) - 1)
	return n
}

func RandFloat32() float32 {
	return rand.Float32()
}

func RandTime() time.Time {
	return time.Unix(int64(RandUint64Exp()), 0)
}

func RandBytes(n int) []byte {
	bs := make([]byte, n)
	for i := 0; i < n; i++ {
		bs[i] = byte(rand.Intn(256))
	}
	return bs
}

// NOTE: This relies on the os's random number generator.
// For real security, we should salt that with some seed.
// See github.com/tendermint/go-crypto for a more secure reader.
func cRandBytes(numBytes int) []byte {
	b := make([]byte, numBytes)
	_, err := crand.Read(b)
	if err != nil {
		sanity.PanicCrisis(err)
	}
	return b
}
