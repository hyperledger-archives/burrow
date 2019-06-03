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

	dbm "github.com/tendermint/tendermint/libs/db"
)

type PrefixDB struct {
	prefix Prefix
	db     dbm.DB
}

func NewPrefixDB(db dbm.DB, prefix string) *PrefixDB {
	return &PrefixDB{
		prefix: Prefix(prefix),
		db:     db,
	}
}

// DB implementation
func (pdb *PrefixDB) Get(key []byte) []byte {
	return pdb.db.Get(pdb.prefix.Key(key))
}

func (pdb *PrefixDB) Has(key []byte) bool {
	return pdb.db.Has(pdb.prefix.Key(key))
}

func (pdb *PrefixDB) Set(key, value []byte) {
	pdb.db.Set(pdb.prefix.Key(key), value)
}

func (pdb *PrefixDB) SetSync(key, value []byte) {
	pdb.db.SetSync(pdb.prefix.Key(key), value)
}

func (pdb *PrefixDB) Delete(key []byte) {
	pdb.db.Delete(pdb.prefix.Key(key))
}

func (pdb *PrefixDB) DeleteSync(key []byte) {
	pdb.db.DeleteSync(pdb.prefix.Key(key))
}

func (pdb *PrefixDB) Iterator(low, high []byte) KVIterator {
	return pdb.prefix.Iterator(pdb.db.Iterator, low, high)
}

func (pdb *PrefixDB) ReverseIterator(low, high []byte) KVIterator {
	return pdb.prefix.Iterator(pdb.db.ReverseIterator, low, high)
}

func (pdb *PrefixDB) Close() {
	pdb.db.Close()
}

func (pdb *PrefixDB) Print() {
	pdb.db.Print()
}

func (pdb *PrefixDB) Stats() map[string]string {
	stats := make(map[string]string)
	stats["PrefixDB.prefix.string"] = string(pdb.prefix)
	stats["PrefixDB.prefix.hex"] = fmt.Sprintf("%X", pdb.prefix)
	source := pdb.db.Stats()
	for key, value := range source {
		stats["PrefixDB.db."+key] = value
	}
	return stats
}

func (pdb *PrefixDB) NewBatch() dbm.Batch {
	return &prefixBatch{
		prefix: pdb.prefix,
		batch:  pdb.db.NewBatch(),
	}
}

type prefixBatch struct {
	prefix Prefix
	batch  dbm.Batch
}

func (pb *prefixBatch) Set(key, value []byte) {
	pb.batch.Set(pb.prefix.Key(key), value)
}

func (pb *prefixBatch) Delete(key []byte) {
	pb.batch.Delete(pb.prefix.Key(key))
}

func (pb *prefixBatch) Write() {
	pb.batch.Write()
}

func (pb *prefixBatch) WriteSync() {
	pb.batch.WriteSync()
}

func (pb *prefixBatch) Close() {
	pb.batch.Close()
}
