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

package storage

import "testing"

func TestNewChannelIterator(t *testing.T) {
	ch := make(chan KVPair)
	go sendKVPair(ch, kvPairs("a", "hello", "b", "channel", "c", "this is nice"))
	ci := NewChannelIterator(ch, bz("a"), bz("c"))
	checkItem(t, ci, bz("a"), bz("hello"))
	checkNext(t, ci, true)
	checkItem(t, ci, bz("b"), bz("channel"))
	checkNext(t, ci, true)
	checkItem(t, ci, bz("c"), bz("this is nice"))
	checkNext(t, ci, false)
	checkInvalid(t, ci)
}
