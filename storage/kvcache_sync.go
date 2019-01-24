package storage

import (
	"sort"
	"sync"
)

type KVCacheSync struct {
	cache sync.Map
}

// Creates an in-memory cache wrapping a map that stores the provided tombstone value for deleted keys
func NewKVCacheSync() *KVCacheSync {
	return &KVCacheSync{
		cache: sync.Map{},
	}
}

func (kvc *KVCacheSync) Info(key []byte) (value []byte, deleted bool) {
	result, ok := kvc.cache.Load(string(key))
	if !ok {
		return nil, false
	}

	vi := result.(valueInfo)
	return vi.value, vi.deleted
}

func (kvc *KVCacheSync) Get(key []byte) []byte {
	result, ok := kvc.cache.Load(string(key))
	if !ok {
		return nil
	}

	vi := result.(valueInfo)
	return vi.value
}

func (kvc *KVCacheSync) Has(key []byte) bool {
	result, ok := kvc.cache.Load(string(key))
	return ok && !result.(valueInfo).deleted
}

func (kvc *KVCacheSync) Set(key, value []byte) {
	skey := string(key)
	vi := valueInfo{
		deleted: false,
		value:   value,
	}
	kvc.cache.Store(skey, vi)
}

func (kvc *KVCacheSync) Delete(key []byte) {
	skey := string(key)
	vi := valueInfo{
		deleted: true,
	}
	kvc.cache.Store(skey, vi)
}

func (kvc *KVCacheSync) Iterator(start, end []byte) KVIterator {
	return kvc.newIterator(NormaliseDomain(start, end, false))
}

func (kvc *KVCacheSync) ReverseIterator(start, end []byte) KVIterator {
	return kvc.newIterator(NormaliseDomain(start, end, true))
}

func (kvc *KVCacheSync) newIterator(start, end []byte) *KVCacheSyncIterator {
	kvi := &KVCacheSyncIterator{
		start: start,
		end:   end,
		keys:  kvc.SortedKeysInDomain(start, end),
		cache: kvc.cache,
	}
	return kvi
}

// Writes contents of cache to backend without flushing the cache
func (kvi *KVCacheSync) WriteTo(writer KVWriter) {
	kvi.cache.Range(func(k, value interface{}) bool {
		kb := []byte(k.(string))
		vi := value.(valueInfo)
		if vi.deleted {
			writer.Delete(kb)
		} else {
			writer.Set(kb, vi.value)
		}
		return true
	})
}

func (kvc *KVCacheSync) Reset() {
	kvc.cache = sync.Map{}
}

type KVCacheSyncIterator struct {
	cache sync.Map
	start []byte
	end   []byte
	keys  [][]byte
	index int
}

func (kvi *KVCacheSyncIterator) Domain() ([]byte, []byte) {
	return kvi.start, kvi.end
}

func (kvi *KVCacheSyncIterator) Info() (key, value []byte, deleted bool) {
	key = kvi.keys[kvi.index]
	result, ok := kvi.cache.Load(string(key))
	if ok {
		vi := result.(valueInfo)
		return key, vi.value, vi.deleted
	} else {
		return key, nil, false
	}
}

func (kvi *KVCacheSyncIterator) Key() []byte {
	return []byte(kvi.keys[kvi.index])
}

func (kvi *KVCacheSyncIterator) Value() []byte {
	result, ok := kvi.cache.Load(string(kvi.keys[kvi.index]))
	if ok {
		return result.(valueInfo).value
	} else {
		return nil
	}
}

func (kvi *KVCacheSyncIterator) Next() {
	if !kvi.Valid() {
		panic("KVCacheSyncIterator.Next() called on invalid iterator")
	}
	kvi.index++
}

func (kvi *KVCacheSyncIterator) Valid() bool {
	return kvi.index < len(kvi.keys)
}

func (kvi *KVCacheSyncIterator) Close() {}

func (kvc *KVCacheSync) SortedKeys(reverse bool) [][]byte {
	keys := make(byteSlices, 0, 0)
	kvc.cache.Range(func(k, value interface{}) bool {
		keys = append(keys, []byte(k.(string)))
		return true
	})
	var sortable sort.Interface = keys
	if reverse {
		sortable = sort.Reverse(keys)
	}
	sort.Stable(sortable)
	return keys
}

func (kvc *KVCacheSync) SortedKeysInDomain(start, end []byte) [][]byte {
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
