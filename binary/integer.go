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

package binary

import (
	"encoding/binary"
	"math"
	"math/big"
	"sort"
)

var big1 = big.NewInt(1)
var tt256 = new(big.Int).Lsh(big1, 256)
var tt256m1 = new(big.Int).Sub(new(big.Int).Lsh(big1, 256), big1)
var tt255 = new(big.Int).Lsh(big1, 255)

// Sort for []uint64

type Uint64Slice []uint64

func (p Uint64Slice) Len() int           { return len(p) }
func (p Uint64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Uint64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p Uint64Slice) Sort()              { sort.Sort(p) }

func SearchUint64s(a []uint64, x uint64) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

func (p Uint64Slice) Search(x uint64) int { return SearchUint64s(p, x) }

//--------------------------------------------------------------------------------

func PutUint64LE(dest []byte, i uint64) {
	binary.LittleEndian.PutUint64(dest, i)
}

func GetUint64LE(src []byte) uint64 {
	return binary.LittleEndian.Uint64(src)
}

func PutUint64BE(dest []byte, i uint64) {
	binary.BigEndian.PutUint64(dest, i)
}

func GetUint64BE(src []byte) uint64 {
	return binary.BigEndian.Uint64(src)
}

func PutInt64LE(dest []byte, i int64) {
	binary.LittleEndian.PutUint64(dest, uint64(i))
}

func GetInt64LE(src []byte) int64 {
	return int64(binary.LittleEndian.Uint64(src))
}

func PutInt64BE(dest []byte, i int64) {
	binary.BigEndian.PutUint64(dest, uint64(i))
}

func GetInt64BE(src []byte) int64 {
	return int64(binary.BigEndian.Uint64(src))
}

// Returns whether a + b would be a uint64 overflow
func IsUint64SumOverflow(a, b uint64) bool {
	return math.MaxUint64-a < b
}

//

// Converts a possibly negative big int x into a positive big int encoding a twos complement representation of x
// truncated to 32 bytes
func U256(x *big.Int) *big.Int {
	// Note that the And operation induces big.Int to hold a positive representation of a negative number
	return new(big.Int).And(x, tt256m1)
}

// Interprets a positive big.Int as a 256-bit two's complement signed integer
func S256(x *big.Int) *big.Int {
	// Sign bit not set, value is its positive self
	if x.Cmp(tt255) < 0 {
		return x
	} else {
		// negative value is represented
		return new(big.Int).Sub(x, tt256)
	}
}

// Treats the positive big int x as if it contains an embedded a back + 1 byte signed integer in its least significant
// bits and extends that sign
func SignExtend(back uint64, x *big.Int) *big.Int {
	// we assume x contains a signed integer of back + 1 bytes width
	// most significant bit of the back'th byte,
	signBit := back*8 + 7
	// single bit set at sign bit position
	mask := new(big.Int).Lsh(big1, uint(signBit))
	// all bits below sign bit set to 1 all above (including sign bit) set to 0
	mask.Sub(mask, big1)
	if x.Bit(int(signBit)) == 1 {
		// Number represented is negative - set all bits above sign bit (including sign bit)
		return x.Or(x, mask.Not(mask))
	} else {
		// Number represented is positive - clear all bits above sign bit (including sign bit)
		return x.And(x, mask)
	}
}
