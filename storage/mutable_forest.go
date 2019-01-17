package storage

import (
	"fmt"

	dbm "github.com/tendermint/tendermint/libs/db"
)

const (
	commitsPrefix = "c"
	treePrefix    = "t"
)

type ForestReader interface {
	Reader(prefix []byte) (KVCallbackIterableReader, error)
}

// MutableForest is a collection of versioned lazily-loaded RWTrees organised by prefix
type MutableForest struct {
	// A tree containing a reference for all contained trees in the form of prefix -> CommitID
	commitsTree *RWTree
	*ImmutableForest
}

func NewMutableForest(db dbm.DB, cacheSize int) (*MutableForest, error) {
	tree := NewRWTree(NewPrefixDB(db, commitsPrefix), cacheSize)
	forest, err := NewImmutableForest(tree, NewPrefixDB(db, treePrefix), cacheSize)
	if err != nil {
		return nil, err
	}
	return &MutableForest{
		ImmutableForest: forest,
		commitsTree:     tree,
	}, nil
}

func (rwf *MutableForest) Load(version int64) error {
	return rwf.commitsTree.Load(version)
}

func (rwf *MutableForest) Save() ([]byte, int64, error) {
	// Save each tree in forest that requires save
	for prefix, tree := range rwf.dirty {
		err := rwf.saveTree([]byte(prefix), tree)
		if err != nil {
			return nil, 0, err
		}
	}
	// empty dirty cache
	rwf.dirty = make(map[string]*RWTree, len(rwf.dirty))
	return rwf.commitsTree.Save()
}

func (rwf *MutableForest) GetImmutable(version int64) (*ImmutableForest, error) {
	commitsTree, err := rwf.commitsTree.GetImmutable(version)
	if err != nil {
		return nil, fmt.Errorf("MutableForest.GetImmutable() could not get commits tree for version %d: %v",
			version, err)
	}
	return NewImmutableForest(commitsTree, rwf.treeDB, rwf.cacheSize)
}

// This function must only be called deterministically (i.e. from EVM and not from RPC layer) since it will cause the returned
// tree to be saved which will alter the hash. Use ReadTree() for read-only queries.
func (rwf *MutableForest) Writer(prefix []byte) (*RWTree, error) {
	tree, err := rwf.tree(prefix)
	if err != nil {
		return nil, err
	}
	rwf.dirty[string(prefix)] = tree
	if tree.Version() == 0 {
		// Ensure tree is available for iteration at genesis version
		rwf.setCommit(prefix, []byte{}, 0)
	}
	return tree, nil
}

// Delete a tree - if the tree exists will return the CommitID of the latest saved version
func (rwf *MutableForest) Delete(prefix []byte) (*CommitID, error) {
	bs, removed := rwf.commitsTree.Delete(prefix)
	if !removed {
		return nil, nil
	}
	return UnmarshalCommitID(bs)
}

// Get the current global hash for all trees in this forest
func (rwf *MutableForest) Hash() []byte {
	return rwf.commitsTree.Hash()
}

// Get the current global version for all versions of all trees in this forest
func (rwf *MutableForest) Version() int64 {
	return rwf.commitsTree.Version()
}

func (rwf *MutableForest) saveTree(prefix []byte, tree *RWTree) error {
	hash, version, err := tree.Save()
	if err != nil {
		return fmt.Errorf("RWForest.saveTree() could not save tree: %v", err)
	}
	return rwf.setCommit(prefix, hash, version)
}

func (rwf *MutableForest) setCommit(prefix, hash []byte, version int64) error {
	bs, err := MarshalCommitID(hash, version)
	if err != nil {
		return fmt.Errorf("RWForest.saveTree() could not marshal CommitID: %v", err)
	}
	rwf.commitsTree.Set([]byte(prefix), bs)
	return nil
}


