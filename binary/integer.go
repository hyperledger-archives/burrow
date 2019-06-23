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
	"math"
	"math/big"
)

var big1 = big.NewInt(1)
var Big256 = big.NewInt(256)
var tt256 = new(big.Int).Lsh(big1, 256)
var tt256m1 = new(big.Int).Sub(new(big.Int).Lsh(big1, 256), big1)
var tt255 = new(big.Int).Lsh(big1, 255)

// Returns whether a + b would be a uint64 overflow
func IsUint64SumOverflow(a, b uint64) bool {
	return math.MaxUint64-a < b
}

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
