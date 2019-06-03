// Copyright 2019 Monax Industries Limited
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

const Word160Length = 20
const Word256Word160Delta = 12

var Zero160 = Word160{}

type Word160 [Word160Length]byte

// Pad a Word160 on the left and embed it in a Word256 (as it is for account addresses in EVM)
func (w Word160) Word256() (word256 Word256) {
	copy(word256[Word256Word160Delta:], w[:])
	return
}

func (w Word160) Bytes() []byte {
	return w[:]
}
