package storage

import (
	"bytes"
	"sort"
	"sync"
)

type KVCache struct {
	sync.RWMutex
	cache map[string]valueInfo
	// Store a sortable slice of keys to avoid always hitting
	keys byteSlices
}

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
	kvc.RLock()
	defer kvc.RUnlock()
	vi := kvc.cache[string(key)]
	return vi.value, vi.deleted
}

func (kvc *KVCache) Get(key []byte) []byte {
	kvc.RLock()
	defer kvc.RUnlock()
	return kvc.cache[string(key)].value
}

func (kvc *KVCache) Has(key []byte) bool {
	kvc.RLock()
	defer kvc.RUnlock()
	vi, ok := kvc.cache[string(key)]
	return ok && !vi.deleted
}

func (kvc *KVCache) Set(key, value []byte) {
	kvc.Lock()
	defer kvc.Unlock()
	skey := string(key)
	vi, ok := kvc.cache[skey]
	if !ok {
		// first Set/Delete
		kvc.keys = append(kvc.keys, key)
		// This slows down write quite a lot but does give faster repeated iterations
		// kvc.keys = insertKey(kvc.keys, key)
	}
	vi.deleted = false
	vi.value = value
	kvc.cache[skey] = vi
}

func (kvc *KVCache) Delete(key []byte) {
	kvc.Lock()
	defer kvc.Unlock()
	skey := string(key)
	vi, ok := kvc.cache[skey]
	if !ok {
		// first Set/Delete
		kvc.keys = append(kvc.keys, key)
		// This slows down write quite a lot but does give faster repeated iterations
		// kvc.keys = insertKey(kvc.keys, key)
	}
	vi.deleted = true
	vi.value = nil
	kvc.cache[skey] = vi
}

func (kvc *KVCache) Iterator(low, high []byte) KVIterator {
	kvc.RLock()
	defer kvc.RUnlock()
	low, high = NormaliseDomain(low, high)
	return kvc.newIterator(low, high, false)
}

func (kvc *KVCache) ReverseIterator(low, high []byte) KVIterator {
	kvc.RLock()
	defer kvc.RUnlock()
	low, high = NormaliseDomain(low, high)
	return kvc.newIterator(low, high, true)
}

// Writes contents of cache to backend without flushing the cache
func (kvc *KVCache) WriteTo(writer KVWriter) {
	kvc.Lock()
	defer kvc.Unlock()
	for k, vi := range kvc.cache {
		kb := []byte(k)
		if vi.deleted {
			writer.Delete(kb)
		} else {
			writer.Set(kb, vi.value)
		}
	}
}

func (kvc *KVCache) Reset() {
	kvc.Lock()
	defer kvc.Unlock()
	kvc.cache = make(map[string]valueInfo)
}

func (kvc *KVCache) sortedKeysInDomain(low, high []byte) [][]byte {
	// Sort keys (which may be partially sorted if we have iterated before)
	sort.Sort(kvc.keys)
	sortedKeys := kvc.keys
	// Attempt to seek to the first key in the range
	startIndex := len(kvc.keys)
	for i, key := range sortedKeys {
		// !(key < start) => key >= start then include (inclusive start)
		if CompareKeys(key, low) != -1 {
			startIndex = i
			break
		}
	}
	// Reslice to beginning of range or end if not found
	sortedKeys = sortedKeys[startIndex:]
	for i, key := range sortedKeys {
		// !(key < end) => key >= end then exclude (exclusive end)
		if CompareKeys(key, high) != -1 {
			sortedKeys = sortedKeys[:i]
			break
		}
	}
	return sortedKeys
}

func (kvc *KVCache) newIterator(start, end []byte, reverse bool) *KVCacheIterator {
	keys := kvc.sortedKeysInDomain(start, end)
	kvi := &KVCacheIterator{
		start:   start,
		end:     end,
		keys:    keys,
		cache:   kvc.cache,
		reverse: reverse,
	}
	return kvi
}

type KVCacheIterator struct {
	cache    map[string]valueInfo
	start    []byte
	end      []byte
	keys     [][]byte
	keyIndex int
	reverse  bool
}

func (kvi *KVCacheIterator) Domain() ([]byte, []byte) {
	return kvi.start, kvi.end
}

func (kvi *KVCacheIterator) Info() (key, value []byte, deleted bool) {
	key = kvi.keys[kvi.sliceIndex()]
	vi := kvi.cache[string(key)]
	return key, vi.value, vi.deleted
}

func (kvi *KVCacheIterator) Key() []byte {
	return []byte(kvi.keys[kvi.sliceIndex()])
}

func (kvi *KVCacheIterator) Value() []byte {
	return kvi.cache[string(kvi.keys[kvi.sliceIndex()])].value
}

func (kvi *KVCacheIterator) Next() {
	if !kvi.Valid() {
		panic("KVCacheIterator.Next() called on invalid iterator")
	}
	kvi.keyIndex++
}

func (kvi *KVCacheIterator) Valid() bool {
	return kvi.keyIndex < len(kvi.keys)
}

func (kvi *KVCacheIterator) Close() {}

func (kvi *KVCacheIterator) sliceIndex() int {
	if kvi.reverse {
		//reflect
		return len(kvi.keys) - 1 - kvi.keyIndex
	}
	return kvi.keyIndex
}

func insertKey(sortedKeys [][]byte, key []byte) [][]byte {
	i := sort.Search(len(sortedKeys), func(i int) bool {
		// Smallest sortedKey such that key
		return bytes.Compare(sortedKeys[i], key) > -1
	})
	// ensure space
	sortedKeys = append(sortedKeys, nil)
	copy(sortedKeys[i+1:], sortedKeys[i:])
	sortedKeys[i] = key
	return sortedKeys
}
