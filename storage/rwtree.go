package storage

import (
	"fmt"
	"unicode/utf8"

	hex "github.com/tmthrgd/go-hex"

	"github.com/tendermint/iavl"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/xlab/treeprint"
)

type RWTree struct {
	// Working tree accumulating writes
	tree *MutableTree
	// Read-only tree serving previous state
	*ImmutableTree
}

// We wrap IAVL's tree types in order to provide iteration helpers and to harmonise other interface types with what we
// expect

type MutableTree struct {
	*iavl.MutableTree
}

func NewMutableTree(db dbm.DB, cacheSize int) *MutableTree {
	tree := iavl.NewMutableTree(db, cacheSize)
	return &MutableTree{
		MutableTree: tree,
	}
}

type ImmutableTree struct {
	*iavl.ImmutableTree
}

func NewRWTree(db dbm.DB, cacheSize int) *RWTree {
	tree := NewMutableTree(db, cacheSize)
	return &RWTree{
		tree: tree,
		// Initially we set readTree to be the inner ImmutableTree of our write tree - this allows us to keep treeVersion == height (FTW)
		ImmutableTree: tree.asImmutable(),
	}
}

// Tries to load the execution state from DB, returns nil with no error if no state found
func (rwt *RWTree) Load(version int64) error {
	const errHeader = "RWTree.Load():"
	if version <= 0 {
		return fmt.Errorf("%s trying to load from non-positive version %d", errHeader, version)
	}
	err := rwt.tree.Load(version)
	if err != nil {
		return fmt.Errorf("%s %v", errHeader, err)
	}
	// Set readTree at commit point == tree
	rwt.ImmutableTree, err = rwt.tree.GetImmutable(version)
	if err != nil {
		return fmt.Errorf("%s %v", errHeader, errHeader)
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
	rwt.ImmutableTree, err = rwt.tree.GetImmutable(version)
	if err != nil {
		return nil, 0, fmt.Errorf("RWTree.Save() could not obtain ImmutableTree read tree: %v", err)
	}
	return hash, version, nil
}

func (rwt *RWTree) Set(key, value []byte) bool {
	return rwt.tree.Set(key, value)
}

func (rwt *RWTree) Delete(key []byte) ([]byte, bool) {
	return rwt.tree.Remove(key)
}

func (rwt *RWTree) GetImmutable(version int64) (*ImmutableTree, error) {
	return rwt.tree.GetImmutable(version)
}

func (rwt *RWTree) IterateWriteTree(start, end []byte, ascending bool, fn func(key []byte, value []byte) error) error {
	return rwt.tree.IterateWriteTree(start, end, ascending, fn)
}

// MutableTree
func (mut *MutableTree) Load(version int64) error {
	if version <= 0 {
		return fmt.Errorf("trying to load MutableTree from non-positive version: version %d", version)
	}
	treeVersion, err := mut.LoadVersionForOverwriting(version)
	if err != nil {
		return fmt.Errorf("could not load current version of MutableTree (version %d): %v", version, err)
	}
	if treeVersion != version {
		return fmt.Errorf("tried to load version %d of MutableTree, but got version %d", version, treeVersion)
	}
	return nil
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

// Get the current working tree as an ImmutableTree (for the methods - not immutable!)
func (mut *MutableTree) asImmutable() *ImmutableTree {
	return &ImmutableTree{mut.MutableTree.ImmutableTree}
}

func (mut *MutableTree) Iterate(start, end []byte, ascending bool, fn func(key []byte, value []byte) error) error {
	return mut.asImmutable().Iterate(start, end, ascending, fn)
}

func (mut *MutableTree) IterateWriteTree(start, end []byte, ascending bool, fn func(key []byte, value []byte) error) error {
	var err error
	mut.MutableTree.IterateRange(start, end, ascending, func(key, value []byte) bool {
		err = fn(key, value)
		if err != nil {
			// stop
			return true
		}
		return false
	})
	return err
}

// ImmutableTree

func (imt *ImmutableTree) Get(key []byte) []byte {
	_, value := imt.ImmutableTree.Get(key)
	return value
}

func (imt *ImmutableTree) Iterate(start, end []byte, ascending bool, fn func(key []byte, value []byte) error) error {
	var err error
	imt.ImmutableTree.IterateRange(start, end, ascending, func(key, value []byte) bool {
		err = fn(key, value)
		if err != nil {
			// stop
			return true
		}
		return false
	})
	return err
}

// Tree printing

func (rwt *RWTree) Dump() string {
	tree := treeprint.New()
	AddTreePrintTree("ReadTree", tree, rwt)
	AddTreePrintTree("WriteTree", tree, rwt.tree)
	return tree.String()
}

func AddTreePrintTree(edge string, tree treeprint.Tree, rwt KVCallbackIterableReader) {
	tree = tree.AddBranch(stringOrHex(edge))
	rwt.Iterate(nil, nil, true, func(key []byte, value []byte) error {
		tree.AddNode(fmt.Sprintf("%s -> %s", stringOrHex(string(key)), stringOrHex(string(value))))
		return nil
	})
}

func stringOrHex(str string) string {
	if utf8.ValidString(str) {
		return str
	}
	return hex.EncodeUpperToString([]byte(str))
}
