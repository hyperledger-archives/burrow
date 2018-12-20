package storage

import (
	"bytes"
	"sort"
)

type KVCache struct {
	cache map[string]valueInfo
}

type valueInfo struct {
	value   []byte
	deleted bool
}

// Creates an in-memory cache wrapping a map that stores the provided tombstone value for deleted keys
func NewKVCache() *KVCache {
	return &KVCache{
		cache: make(map[string]valueInfo),
	}
}

func (kvc *KVCache) Info(key []byte) (value []byte, deleted bool) {
	vi := kvc.cache[string(key)]
	return vi.value, vi.deleted
}

func (kvc *KVCache) Get(key []byte) []byte {
	return kvc.cache[string(key)].value
}

func (kvc *KVCache) Has(key []byte) bool {
	vi, ok := kvc.cache[string(key)]
	return ok && !vi.deleted
}

func (kvc *KVCache) Set(key, value []byte) {
	skey := string(key)
	vi := kvc.cache[skey]
	vi.deleted = false
	vi.value = value
	kvc.cache[skey] = vi
}

func (kvc *KVCache) Delete(key []byte) {
	skey := string(key)
	vi := kvc.cache[skey]
	vi.deleted = true
	vi.value = nil
	kvc.cache[skey] = vi
}

func (kvc *KVCache) Iterator(start, end []byte) KVIterator {
	return kvc.newIterator(NormaliseDomain(start, end, false))
}

func (kvc *KVCache) ReverseIterator(start, end []byte) KVIterator {
	return kvc.newIterator(NormaliseDomain(start, end, true))
}

func (kvc *KVCache) newIterator(start, end []byte) *KVCacheIterator {
	kvi := &KVCacheIterator{
		start: start,
		end:   end,
		keys:  kvc.SortedKeysInDomain(start, end),
		cache: kvc.cache,
	}
	return kvi
}

// Writes contents of cache to backend without flushing the cache
func (kvi *KVCache) WriteTo(writer KVWriter) {
	for k, vi := range kvi.cache {
		kb := []byte(k)
		if vi.deleted {
			writer.Delete(kb)
		} else {
			writer.Set(kb, vi.value)
		}
	}
}

func (kvc *KVCache) Reset() {
	kvc.cache = make(map[string]valueInfo)
}

type KVCacheIterator struct {
	cache map[string]valueInfo
	start []byte
	end   []byte
	keys  [][]byte
	index int
}

func (kvi *KVCacheIterator) Domain() ([]byte, []byte) {
	return kvi.start, kvi.end
}

func (kvi *KVCacheIterator) Info() (key, value []byte, deleted bool) {
	key = kvi.keys[kvi.index]
	vi := kvi.cache[string(key)]
	return key, vi.value, vi.deleted
}

func (kvi *KVCacheIterator) Key() []byte {
	return []byte(kvi.keys[kvi.index])
}

func (kvi *KVCacheIterator) Value() []byte {
	return kvi.cache[string(kvi.keys[kvi.index])].value
}

func (kvi *KVCacheIterator) Next() {
	if !kvi.Valid() {
		panic("KVCacheIterator.Next() called on invalid iterator")
	}
	kvi.index++
}

func (kvi *KVCacheIterator) Valid() bool {
	return kvi.index < len(kvi.keys)
}

func (kvi *KVCacheIterator) Close() {}

type byteSlices [][]byte

func (bss byteSlices) Len() int {
	return len(bss)
}

func (bss byteSlices) Less(i, j int) bool {
	return bytes.Compare(bss[i], bss[j]) == -1
}

func (bss byteSlices) Swap(i, j int) {
	bss[i], bss[j] = bss[j], bss[i]
}

func (kvc *KVCache) SortedKeys(reverse bool) [][]byte {
	keys := make(byteSlices, 0, len(kvc.cache))
	for k := range kvc.cache {
		keys = append(keys, []byte(k))
	}
	var sortable sort.Interface = keys
	if reverse {
		sortable = sort.Reverse(keys)
	}
	sort.Stable(sortable)
	return keys
}

func (kvc *KVCache) SortedKeysInDomain(start, end []byte) [][]byte {
	comp := CompareKeys(start, end)
	if comp == 0 {
		return [][]byte{}
	}
	// Sort keys depending on order of end points
	sortedKeys := kvc.SortedKeys(comp == 1)
	// Attempt to seek to the first key in the range
	startIndex := len(sortedKeys)
	for i, key := range sortedKeys {
		if CompareKeys(key, start) != comp {
			startIndex = i
			break
		}
	}
	// Reslice to beginning of range or end if not found
	sortedKeys = sortedKeys[startIndex:]
	for i, key := range sortedKeys {
		if CompareKeys(key, end) != comp {
			sortedKeys = sortedKeys[:i]
			break
		}
	}
	return sortedKeys
}
