package storage

import (
	"fmt"
	"sync"

	"github.com/tendermint/iavl"
	dbm "github.com/tendermint/tendermint/libs/db"
)

type RWTree struct {
	// Values not reassigned
	sync.RWMutex
	// Working tree accumulating writes
	tree *iavl.MutableTree
	// Read tree serving previous state
	readTree *iavl.ImmutableTree
}

func NewRWTree(db dbm.DB, cacheSize int) *RWTree {
	return &RWTree{
		tree:     iavl.NewMutableTree(db, cacheSize),
		readTree: iavl.NewImmutableTree(db, cacheSize),
	}
}

// Tries to load the execution state from DB, returns nil with no error if no state found
func (rwt *RWTree) Load(version int64) error {
	if version <= 0 {
		return fmt.Errorf("trying to load RWTree from non-positive version: version %d", version)
	}
	treeVersion, err := rwt.tree.LoadVersionForOverwriting(version)
	if err != nil {
		return fmt.Errorf("could not load current version of RWTree: version %d", version)
	}
	if treeVersion != version {
		return fmt.Errorf("tried to load version %d of RWTree, but got version %d", version, treeVersion)
	}
	// Set readTree at commit point == tree
	rwt.readTree, err = rwt.tree.GetImmutable(version)
	if err != nil {
		return fmt.Errorf("could not load previous version of RWTree to use as read version")
	}
	return nil
}

// Save the current write tree making writes accessible from read tree.
func (rwt *RWTree) Save() ([]byte, int64, error) {
	// save state at a new version may still be orphaned before we save the version against the hash
	hash, version, err := rwt.tree.SaveVersion()
	if err != nil {
		return nil, 0, fmt.Errorf("could not save RWTree: %v", err)
	}
	// Take an immutable reference to the tree we just saved for querying
	rwt.readTree, err = rwt.tree.GetImmutable(version)
	if err != nil {
		return nil, 0, fmt.Errorf("RWTree.Save() could not obtain ImmutableTree read tree: %v", err)
	}
	return hash, version, nil
}

func (rwt *RWTree) Set(key, value []byte) bool {
	return rwt.tree.Set(key, value)
}

func (rwt *RWTree) Get(key []byte) []byte {
	_, value := rwt.readTree.Get(key)
	return value
}

func (rwt *RWTree) IterateRange(start, end []byte, ascending bool, fn func(key []byte, value []byte) bool) (stopped bool) {
	return rwt.readTree.IterateRange(start, end, ascending, fn)
}

func (rwt *RWTree) Hash() []byte {
	return rwt.readTree.Hash()
}

func (rwt *RWTree) Has(key []byte) bool {
	return rwt.Get(key) != nil
}

func (rwt *RWTree) Delete(key []byte) ([]byte, bool) {
	return rwt.tree.Remove(key)
}

func (rwt *RWTree) Iterator(start, end []byte) dbm.Iterator {
	ch := make(chan KVPair)
	go func() {
		defer close(ch)
		rwt.readTree.IterateRange(start, end, true, func(key, value []byte) (stop bool) {
			ch <- KVPair{key, value}
			return
		})
	}()
	return NewChannelIterator(ch, start, end)
}

func (rwt *RWTree) ReverseIterator(start, end []byte) dbm.Iterator {
	ch := make(chan KVPair)
	go func() {
		defer close(ch)
		rwt.readTree.IterateRange(start, end, false, func(key, value []byte) (stop bool) {
			ch <- KVPair{key, value}
			return
		})
	}()
	return NewChannelIterator(ch, start, end)
}
