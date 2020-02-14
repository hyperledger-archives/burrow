package storage

import (
	"bytes"
	"fmt"
	"sort"
)

type ChannelIterator struct {
	ch      <-chan KVPair
	start   []byte
	end     []byte
	kv      KVPair
	invalid bool
}

type KVPair struct {
	Key   []byte
	Value []byte
}

func (kv KVPair) String() string {
	return fmt.Sprintf("%s => %s", string(kv.Key), string(kv.Value))
}

type KVPairs []KVPair

func (kvp KVPairs) Len() int {
	return len(kvp)
}

func (kvp KVPairs) Less(i, j int) bool {
	return bytes.Compare(kvp[i].Key, kvp[j].Key) == -1
}

func (kvp KVPairs) Swap(i, j int) {
	kvp[i], kvp[j] = kvp[j], kvp[i]
}

func (kvp KVPairs) Sorted() KVPairs {
	kvpCopy := make(KVPairs, len(kvp))
	copy(kvpCopy, kvp)
	sort.Stable(kvpCopy)
	return kvpCopy
}

// ChannelIterator wraps a stream of kvp KVPairs over a channel as a stateful KVIterator. The start and end keys provided
// are purely indicative (for Domain()) and are assumed to be honoured by the input channel - they are not checked
// and keys are not sorted. NewChannelIterator will block until the first value is received over the channel.
func NewChannelIterator(ch <-chan KVPair, start, end []byte) *ChannelIterator {
	ci := &ChannelIterator{
		ch:    ch,
		start: start,
		end:   end,
	}
	// Load first element if it exists
	ci.Next()
	return ci
}

func (it *ChannelIterator) Domain() ([]byte, []byte) {
	return it.start, it.end
}

func (it *ChannelIterator) Valid() bool {
	return !it.invalid
}

func (it *ChannelIterator) Next() {
	if it.invalid {
		panic("ChannelIterator.Value() called on invalid iterator")
	}
	kv, ok := <-it.ch
	it.invalid = !ok
	it.kv = kv
}

func (it *ChannelIterator) Key() []byte {
	if it.invalid {
		panic("ChannelIterator.Key() called on invalid iterator")
	}
	return it.kv.Key
}

func (it *ChannelIterator) Value() []byte {
	if it.invalid {
		panic("ChannelIterator.Value() called on invalid iterator")
	}
	return it.kv.Value
}

func (it *ChannelIterator) Close() {
	for range it.ch {
		// drain channel if necessary
	}
}

func (it *ChannelIterator) Error() error {
	for range it.ch {
		if err := it.Error(); err != nil {
			return nil
		}
	}
	return nil
}
