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
