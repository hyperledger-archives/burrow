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

import (
	"bytes"

	dbm "github.com/tendermint/tendermint/libs/db"
	hex "github.com/tmthrgd/go-hex"
)

type Prefix []byte

func NewPrefix(bs []byte) Prefix {
	p := make(Prefix, len(bs))
	copy(p, bs)
	return p
}

func (p Prefix) Key(key []byte) []byte {
	// Avoid any unintended memory sharing between keys
	return append(p[:len(p):len(p)], key...)
}

func (p Prefix) Suffix(key []byte) []byte {
	bs := make([]byte, len(key)-len(p))
	copy(bs, key[len(p):])
	return bs
}

// Get the lexicographical sibling above this prefix (i.e. the fixed length integer plus one)
func (p Prefix) Above() []byte {
	for i := len(p) - 1; i >= 0; i-- {
		c := p[i]
		if c < 0xff {
			inc := make([]byte, i+1)
			copy(inc, p)
			inc[i]++
			return inc
		}
	}
	return nil
}

// Get the lexicographical sibling below this prefix (i.e. the fixed length integer minus one)
func (p Prefix) Below() []byte {
	for i := len(p) - 1; i >= 0; i-- {
		c := p[i]
		if c > 0x00 {
			inc := make([]byte, i+1)
			copy(inc, p)
			inc[i]--
			return inc
		}
	}
	return nil
}

func (p Prefix) CallbackIterable(source KVCallbackIterable) *prefixCallbackIterable {
	return &prefixCallbackIterable{
		prefix: p,
		source: source,
	}
}

type prefixCallbackIterable struct {
	prefix Prefix
	source KVCallbackIterable
}

func (pi *prefixCallbackIterable) Iterate(start, end []byte, ascending bool, fn func(key []byte, value []byte) error) error {
	var pstart, pend []byte = pi.prefix.Key(start), nil

	if start == nil {
		// We may iterate on a key that does not start with prefix
		pstart = pi.prefix.Below()
	} else {
		pstart = pi.prefix.Key(start)
	}
	if end == nil {
		// Source is exclusive on end so we won't iterate over it
		pend = pi.prefix.Above()
	} else {
		pend = pi.prefix.Key(end)
	}
	return pi.source.Iterate(pstart, pend, ascending, func(key []byte, value []byte) error {
		if bytes.HasPrefix(key, pi.prefix) {
			return fn(pi.prefix.Suffix(key), value)
		}
		return nil
	})
}

func (p Prefix) Iterator(iteratorFn func(start, end []byte) dbm.Iterator, start, end []byte) KVIterator {
	var pstart, pend []byte = p.Key(start), nil

	if end == nil {
		pend = p.Above()
	} else {
		pend = p.Key(end)
	}
	return &prefixIterator{
		start:  start,
		end:    end,
		prefix: p,
		source: iteratorFn(pstart, pend),
	}
}

func (p Prefix) Iterable(source KVIterable) KVIterable {
	return &prefixIterable{
		prefix: p,
		source: source,
	}
}

type prefixIterable struct {
	prefix Prefix
	source KVIterable
}

func (pi *prefixIterable) Iterator(low, high []byte) KVIterator {
	return pi.prefix.Iterator(pi.source.Iterator, low, high)
}

func (pi *prefixIterable) ReverseIterator(low, high []byte) KVIterator {
	return pi.prefix.Iterator(pi.source.ReverseIterator, low, high)
}

func (p Prefix) Store(source KVStore) KVStore {
	return &prefixKVStore{
		prefix: p,
		source: source,
	}
}

func (p Prefix) Length() int {
	return len(p)
}

func (p Prefix) String() string {
	return string(p)
}

func (p Prefix) HexString() string {
	return hex.EncodeUpperToString(p)
}

type prefixIterator struct {
	prefix  Prefix
	source  dbm.Iterator
	start   []byte
	end     []byte
	invalid bool
}

func (pi *prefixIterator) Domain() ([]byte, []byte) {
	return pi.start, pi.end
}

func (pi *prefixIterator) Valid() bool {
	pi.validate()
	return !pi.invalid && pi.source.Valid()
}

func (pi *prefixIterator) Next() {
	if pi.invalid {
		panic("prefixIterator.Next() called on invalid iterator")
	}
	pi.source.Next()
	pi.validate()
}

func (pi *prefixIterator) Key() []byte {
	if pi.invalid {
		panic("prefixIterator.Key() called on invalid iterator")
	}
	return pi.prefix.Suffix(pi.source.Key())
}

func (pi *prefixIterator) Value() []byte {
	if pi.invalid {
		panic("prefixIterator.Value() called on invalid iterator")
	}
	return pi.source.Value()
}

func (pi *prefixIterator) Close() {
	pi.source.Close()
}

func (pi *prefixIterator) validate() {
	if pi.invalid {
		return
	}
	sourceValid := pi.source.Valid()
	pi.invalid = !sourceValid || !bytes.HasPrefix(pi.source.Key(), pi.prefix)
	if pi.invalid {
		pi.Close()
	}
}

type prefixKVStore struct {
	prefix Prefix
	source KVStore
}

func (ps *prefixKVStore) Get(key []byte) []byte {
	return ps.source.Get(ps.prefix.Key(key))
}

func (ps *prefixKVStore) Has(key []byte) bool {
	return ps.source.Has(ps.prefix.Key(key))
}

func (ps *prefixKVStore) Set(key, value []byte) {
	ps.source.Set(ps.prefix.Key(key), value)
}

func (ps *prefixKVStore) Delete(key []byte) {
	ps.source.Delete(ps.prefix.Key(key))
}

func (ps *prefixKVStore) Iterator(low, high []byte) KVIterator {
	return ps.prefix.Iterator(ps.source.Iterator, low, high)
}

func (ps *prefixKVStore) ReverseIterator(low, high []byte) KVIterator {
	return ps.prefix.Iterator(ps.source.ReverseIterator, low, high)
}
