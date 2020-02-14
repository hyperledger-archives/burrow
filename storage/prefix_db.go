package storage

import (
	"fmt"

	dbm "github.com/tendermint/tm-db"
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
func (pdb *PrefixDB) Get(key []byte) ([]byte, error) {
	return pdb.db.Get(pdb.prefix.Key(key))
}

func (pdb *PrefixDB) Has(key []byte) (bool, error) {
	return pdb.db.Has(pdb.prefix.Key(key))
}

func (pdb *PrefixDB) Set(key, value []byte) error {
	return pdb.db.Set(pdb.prefix.Key(key), value)
}

func (pdb *PrefixDB) SetSync(key, value []byte) error {
	return pdb.db.SetSync(pdb.prefix.Key(key), value)
}

func (pdb *PrefixDB) Delete(key []byte) error {
	return pdb.db.Delete(pdb.prefix.Key(key))
}

func (pdb *PrefixDB) DeleteSync(key []byte) error {
	return pdb.db.DeleteSync(pdb.prefix.Key(key))
}

func (pdb *PrefixDB) Iterator(low, high []byte) (KVIterator, error) {
	return pdb.prefix.Iterator(pdb.db.Iterator, low, high)
}

func (pdb *PrefixDB) ReverseIterator(low, high []byte) (KVIterator, error) {
	return pdb.prefix.Iterator(pdb.db.ReverseIterator, low, high)
}

func (pdb *PrefixDB) Close() error {
	return pdb.db.Close()
}

func (pdb *PrefixDB) Print() error {
	return pdb.db.Print()
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

func (pb *prefixBatch) Write() error {
	return pb.batch.Write()
}

func (pb *prefixBatch) WriteSync() error {
	return pb.batch.WriteSync()
}

func (pb *prefixBatch) Close() {
	pb.batch.Close()
}
