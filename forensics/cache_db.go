package forensics

import (
	"github.com/hyperledger/burrow/storage"
	dbm "github.com/tendermint/tendermint/libs/db"
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
func (cdb *CacheDB) Get(key []byte) []byte {
	value, deleted := cdb.cache.Info(key)
	if deleted {
		return nil
	}
	if value != nil {
		return value
	}
	return cdb.backend.Get(key)
}

func (cdb *CacheDB) Has(key []byte) bool {
	value, deleted := cdb.cache.Info(key)
	return !deleted && (value != nil || cdb.backend.Has(key))
}

func (cdb *CacheDB) Iterator(low, high []byte) storage.KVIterator {
	// Keys from cache will sort first because of order in MultiIterator and Uniq will take the first KVs so KVs
	// appearing in cache will override values from backend.
	return Uniq(NewMultiIterator(false, cdb.cache.Iterator(low, high), cdb.backend.Iterator(low, high)))
}

func (cdb *CacheDB) ReverseIterator(low, high []byte) storage.KVIterator {
	return Uniq(NewMultiIterator(true, cdb.cache.ReverseIterator(low, high), cdb.backend.ReverseIterator(low, high)))
}

func (cdb *CacheDB) Set(key, value []byte) {
	cdb.cache.Set(key, value)
}

func (cdb *CacheDB) SetSync(key, value []byte) {
	cdb.cache.Set(key, value)
}

func (cdb *CacheDB) Delete(key []byte) {
	cdb.cache.Delete(key)
}

func (cdb *CacheDB) DeleteSync(key []byte) {
	cdb.Delete(key)
}

func (cdb *CacheDB) Close() {
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

func (cb *cacheBatch) Write() {
	cb.cache.WriteTo(cb.backend)
}

func (cb *cacheBatch) Close() {
}

func (cb *cacheBatch) WriteSync() {
	cb.Write()
}

func (cdb *CacheDB) Print() {
}

func (cdb *CacheDB) Stats() map[string]string {
	return map[string]string{}
}
