package storage

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	dbm "github.com/tendermint/tendermint/libs/db"
)

func TestBatchCommit(t *testing.T) {
	db := dbm.NewMemDB()
	cdb := NewCacheDB(db)
	foo := bz("foo")
	bam := bz("bam")
	bosh := bz("bosh")
	boom := bz("boom")
	db.Set(foo, bam)
	assert.Equal(t, bam, cdb.Get(foo), "underlying writes should be seen")
	cdb.Set(foo, bosh)
	assert.Equal(t, bosh, cdb.Get(foo), "writes to CacheDB should be available")
	batch := cdb.NewBatch()
	batch.Set(foo, bam)
	assert.Equal(t, bosh, cdb.Get(foo), "write to batch should not be seen")
	batch.WriteSync()
	cdb.Commit(db)
	assert.Equal(t, bam, db.Get(foo), "changes should commit")
	cdb.Set(foo, bosh)
	assert.Equal(t, bam, db.Get(foo), "uncommitted changes should not be seen in db")
	cdb.Delete(foo)
	assert.Nil(t, cdb.Get(foo))
	assert.Equal(t, bam, db.Get(foo))
	cdb.Commit(db)
	assert.Nil(t, db.Get(foo))
	cdb.Set(foo, boom)
	assert.Nil(t, db.Get(foo))
}

func TestCacheDB_Iterator(t *testing.T) {
	db := dbm.NewMemDB()
	cdb := NewCacheDB(db)
	foo := bz("foo")
	bam := bz("bam")
	bosh := bz("bosh")
	boom := bz("boom")

	db.Set(append(foo, foo...), foo)
	db.Set(append(foo, bam...), bam)
	cdb.Set(append(foo, bosh...), bosh)
	cdb.Set(boom, boom)

	it := cdb.Iterator(nil, nil)
	kvp := collectIterator(it)
	fmt.Println(kvp)
	cdb.Commit(db)

	it = db.Iterator(nil, nil)
	assert.Equal(t, kvp, collectIterator(it))
}
