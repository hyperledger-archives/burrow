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
	"bytes"
	"fmt"
	"math/big"
	"sort"

	"github.com/tmthrgd/go-hex"
)

var (
	Zero256 = Word256{}
	One256  = LeftPadWord256([]byte{1})
)

const Word256Length = 32

var BigWord256Length = big.NewInt(Word256Length)

var trimCutSet = string([]byte{0})

type Word256 [Word256Length]byte

func (w *Word256) UnmarshalText(hexBytes []byte) error {
	bs, err := hex.DecodeString(string(hexBytes))
	if err != nil {
		return err
	}
	copy(w[:], bs)
	return nil
}

func (w Word256) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeUpperToString(w[:])), nil
}

func (w Word256) String() string {
	return string(w[:])
}

func (w Word256) Copy() Word256 {
	return w
}

func (w Word256) Bytes() []byte {
	return w[:]
}

// copied.
func (w Word256) Prefix(n int) []byte {
	return w[:n]
}

func (w Word256) Postfix(n int) []byte {
	return w[32-n:]
}

// Get a Word160 embedded a Word256 and padded on the left (as it is for account addresses in EVM)
func (w Word256) Word160() (w160 Word160) {
	copy(w160[:], w[Word256Word160Delta:])
	return
}

func (w Word256) IsZero() bool {
	accum := byte(0)
	for _, byt := range w {
		accum |= byt
	}
	return accum == 0
}

func (w Word256) Compare(other Word256) int {
	return bytes.Compare(w[:], other[:])
}

func (w Word256) UnpadLeft() []byte {
	return bytes.TrimLeft(w[:], trimCutSet)
}

func (w Word256) UnpadRight() []byte {
	return bytes.TrimRight(w[:], trimCutSet)
}

// Gogo proto support
func (w *Word256) Marshal() ([]byte, error) {
	if w == nil {
		return nil, nil
	}
	return w.Bytes(), nil
}

func (w *Word256) Unmarshal(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if len(data) != Word256Length {
		return fmt.Errorf("error unmarshallling Word256 '%X' from bytes: %d bytes but should have %d bytes",
			data, len(data), Word256Length)
	}
	copy(w[:], data)
	return nil
}

func (w *Word256) MarshalTo(data []byte) (int, error) {
	if w == nil {
		return 0, nil
	}
	return copy(data, w[:]), nil
}

func (w Word256) Size() int {
	return Word256Length
}

func Uint64ToWord256(i uint64) (word Word256) {
	PutUint64BE(word[24:], i)
	return
}

func Int64ToWord256(i int64) (word Word256) {
	PutInt64BE(word[24:], i)
	return
}

func RightPadWord256(bz []byte) (word Word256) {
	copy(word[:], bz)
	return
}

func LeftPadWord256(bz []byte) (word Word256) {
	copy(word[32-len(bz):], bz)
	return
}

func Uint64FromWord256(word Word256) uint64 {
	return GetUint64BE(word.Postfix(8))
}

func Int64FromWord256(word Word256) int64 {
	return GetInt64BE(word.Postfix(8))
}

//-------------------------------------

type Words256 []Word256

func (ws Words256) Len() int {
	return len(ws)
}

func (ws Words256) Less(i, j int) bool {
	return ws[i].Compare(ws[j]) < 0
}

func (ws Words256) Swap(i, j int) {
	ws[i], ws[j] = ws[j], ws[i]
}

type Tuple256 struct {
	First  Word256
	Second Word256
}

func (tuple Tuple256) Compare(other Tuple256) int {
	firstCompare := tuple.First.Compare(other.First)
	if firstCompare == 0 {
		return tuple.Second.Compare(other.Second)
	} else {
		return firstCompare
	}
}

func Tuple256Split(t Tuple256) (Word256, Word256) {
	return t.First, t.Second
}

type Tuple256Slice []Tuple256

func (p Tuple256Slice) Len() int { return len(p) }
func (p Tuple256Slice) Less(i, j int) bool {
	return p[i].Compare(p[j]) < 0
}
func (p Tuple256Slice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p Tuple256Slice) Sort()         { sort.Sort(p) }
