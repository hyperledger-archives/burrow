package storage

type KVCascade []KVIterableReader

func (kvc KVCascade) Get(key []byte) []byte {
	for _, kvs := range kvc {
		value := kvs.Get(key)
		if value != nil {
			return value
		}
	}
	return nil
}

func (kvc KVCascade) Has(key []byte) bool {
	for _, kvs := range kvc {
		has := kvs.Has(key)
		if has {
			return true
		}
	}
	return false
}

func (kvc KVCascade) Iterator(start, end []byte) KVIterator {
	iterators := make([]KVIterator, len(kvc))
	for i, kvs := range kvc {
		iterators[i] = kvs.Iterator(start, end)
	}
	return NewMultiIterator(false, iterators...)
}

func (kvc KVCascade) ReverseIterator(start, end []byte) KVIterator {
	iterators := make([]KVIterator, len(kvc))
	for i, kvs := range kvc {
		iterators[i] = kvs.ReverseIterator(start, end)
	}
	return NewMultiIterator(true, iterators...)
}
