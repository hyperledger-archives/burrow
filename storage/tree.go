package storage

import (
	"fmt"

	"github.com/tendermint/iavl"
	dbm "github.com/tendermint/tendermint/libs/db"
)

// We wrap IAVL's tree types in order to implement standard DB interface and iteration helpers
type MutableTree struct {
	*iavl.MutableTree
}

type ImmutableTree struct {
	*iavl.ImmutableTree
}

func NewMutableTree(db dbm.DB, cacheSize int) *MutableTree {
	tree := iavl.NewMutableTree(db, cacheSize)
	return &MutableTree{
		MutableTree: tree,
	}
}

func (mut *MutableTree) Load(version int64, overwriting bool) error {
	if version <= 0 {
		return fmt.Errorf("trying to load MutableTree from non-positive version: version %d", version)
	}
	var err error
	var treeVersion int64
	if overwriting {
		// Deletes all version above version!
		treeVersion, err = mut.MutableTree.LoadVersionForOverwriting(version)
	} else {
		treeVersion, err = mut.MutableTree.LoadVersion(version)
	}
	if err != nil {
		return fmt.Errorf("could not load current version of MutableTree (version %d): %v", version, err)
	}
	if treeVersion != version {
		return fmt.Errorf("tried to load version %d of MutableTree, but got version %d", version, treeVersion)
	}
	return nil
}

func (mut *MutableTree) Iterate(start, end []byte, ascending bool, fn func(key []byte, value []byte) error) error {
	return mut.asImmutable().Iterate(start, end, ascending, fn)
}

func (mut *MutableTree) IterateWriteTree(start, end []byte, ascending bool, fn func(key []byte, value []byte) error) error {
	var err error
	mut.MutableTree.IterateRange(start, end, ascending, func(key, value []byte) (stop bool) {
		err = fn(key, value)
		return err != nil
	})
	return err
}

func (mut *MutableTree) Get(key []byte) []byte {
	_, bs := mut.MutableTree.Get(key)
	return bs
}

func (mut *MutableTree) GetImmutable(version int64) (*ImmutableTree, error) {
	tree, err := mut.MutableTree.GetImmutable(version)
	if err != nil {
		return nil, err
	}
	return &ImmutableTree{tree}, nil
}

func (imt *ImmutableTree) Get(key []byte) []byte {
	_, value := imt.ImmutableTree.Get(key)
	return value
}

func (imt *ImmutableTree) Iterate(start, end []byte, ascending bool, fn func(key []byte, value []byte) error) error {
	var err error
	imt.ImmutableTree.IterateRange(start, end, ascending, func(key, value []byte) bool {
		err = fn(key, value)
		return err != nil
	})
	return err
}

// Get the current working tree as an ImmutableTree (for the methods - not immutable!)
func (mut *MutableTree) asImmutable() *ImmutableTree {
	return &ImmutableTree{mut.MutableTree.ImmutableTree}
}
