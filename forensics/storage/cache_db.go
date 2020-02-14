package storage

import (
	"github.com/hyperledger/burrow/storage"
	dbm "github.com/tendermint/tm-db"
)

type CacheDB struct {
	cache   *KVCache
	backend storage.KVIterableReader
}

func NewCacheDB(backend storage.KVIterableReader) *CacheDB {
	return &CacheDB{
		cache:   NewKVCache(),
		backend: backend,
	}
}

// DB implementation
func (cdb *CacheDB) Get(key []byte) ([]byte, error) {
	value, deleted := cdb.cache.Info(key)
	if deleted {
		return nil, nil
	}
	if value != nil {
		return value, nil
	}
	return cdb.backend.Get(key)
}

func (cdb *CacheDB) Has(key []byte) (bool, error) {
	value, deleted := cdb.cache.Info(key)
	has, err := cdb.backend.Has(key)
	if err != nil {
		return false, err
	}
	return !deleted && (value != nil || has), nil
}

func (cdb *CacheDB) Iterator(low, high []byte) (storage.KVIterator, error) {
	// Keys from cache will sort first because of order in MultiIterator and Uniq will take the first KVs so KVs
	// appearing in cache will override values from backend.
	iterator, err := cdb.backend.Iterator(low, high)
	if err != nil {
		return nil, err
	}

	return Uniq(NewMultiIterator(false, cdb.cache.Iterator(low, high), iterator)), nil
}

func (cdb *CacheDB) ReverseIterator(low, high []byte) (storage.KVIterator, error) {
	iterator, err := cdb.backend.ReverseIterator(low, high)
	if err != nil {
		return nil, err
	}

	return Uniq(NewMultiIterator(true, cdb.cache.ReverseIterator(low, high), iterator)), nil
}

func (cdb *CacheDB) Set(key, value []byte) error {
	cdb.cache.Set(key, value)
	return nil
}

func (cdb *CacheDB) SetSync(key, value []byte) error {
	cdb.cache.Set(key, value)
	return nil
}

func (cdb *CacheDB) Delete(key []byte) error {
	cdb.cache.Delete(key)
	return nil
}

func (cdb *CacheDB) DeleteSync(key []byte) error {
	return cdb.Delete(key)
}

func (cdb *CacheDB) Close() error {
	return nil
}

func (cdb *CacheDB) NewBatch() dbm.Batch {
	return &cacheBatch{
		cache:   NewKVCache(),
		backend: cdb,
	}
}

func (cdb *CacheDB) Commit(writer storage.KVWriter) {
	cdb.cache.WriteTo(writer)
	cdb.cache.Reset()
}

type cacheBatch struct {
	cache   *KVCache
	backend *CacheDB
}

func (cb *cacheBatch) Set(key, value []byte) {
	cb.cache.Set(key, value)
}

func (cb *cacheBatch) Delete(key []byte) {
	cb.cache.Delete(key)
}

func (cb *cacheBatch) Write() error {
	cb.cache.WriteTo(cb.backend)
	return nil
}

func (cb *cacheBatch) Close() {
}

func (cb *cacheBatch) WriteSync() error {
	return cb.Write()
}

func (cdb *CacheDB) Print() error {
	return nil
}

func (cdb *CacheDB) Stats() map[string]string {
	return map[string]string{}
}
