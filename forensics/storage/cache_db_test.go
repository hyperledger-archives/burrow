package storage

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	dbm "github.com/tendermint/tm-db"
)

func TestBatchCommit(t *testing.T) {
	db := dbm.NewMemDB()
	cdb := NewCacheDB(db)
	foo := []byte("foo")
	bam := []byte("bam")
	bosh := []byte("bosh")
	boom := []byte("boom")

	db.Set(foo, bam)
	result, err := cdb.Get(foo)
	assert.NoError(t, err)
	assert.Equal(t, bam, result, "underlying writes should be seen")
	cdb.Set(foo, bosh)
	result, err = cdb.Get(foo)
	assert.NoError(t, err)
	assert.Equal(t, bosh, result, "writes to CacheDB should be available")
	batch := cdb.NewBatch()
	batch.Set(foo, bam)
	result, err = cdb.Get(foo)
	assert.NoError(t, err)
	assert.Equal(t, bosh, result, "write to batch should not be seen")
	batch.WriteSync()
	cdb.Commit(db)
	result, err = db.Get(foo)
	assert.NoError(t, err)
	assert.Equal(t, bam, result, "changes should commit")
	cdb.Set(foo, bosh)
	result, err = db.Get(foo)
	assert.NoError(t, err)
	assert.Equal(t, bam, result, "uncommitted changes should not be seen in db")
	cdb.Delete(foo)
	result, err = cdb.Get(foo)
	assert.NoError(t, err)
	assert.Nil(t, result)
	result, err = db.Get(foo)
	assert.NoError(t, err)
	assert.Equal(t, bam, result)
	cdb.Commit(db)
	result, err = db.Get(foo)
	assert.NoError(t, err)

	assert.Nil(t, result)
	cdb.Set(foo, boom)
	result, err = db.Get(foo)
	assert.NoError(t, err)

	assert.Nil(t, result)
}

func TestCacheDB_Iterator(t *testing.T) {
	db := dbm.NewMemDB()
	cdb := NewCacheDB(db)
	foo := []byte("foo")
	bam := []byte("bam")
	bosh := []byte("bosh")
	boom := []byte("boom")

	db.Set(append(foo, foo...), foo)
	db.Set(append(foo, bam...), bam)
	cdb.Set(append(foo, bosh...), bosh)
	cdb.Set(boom, boom)

	it, err := cdb.Iterator(nil, nil)
	assert.NoError(t, err)
	kvp := collectIterator(it)
	fmt.Println(kvp)
	cdb.Commit(db)

	it, err = db.Iterator(nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, kvp, collectIterator(it))
}
