package storage

import (
	"fmt"

	"github.com/tendermint/iavl"

	dbm "github.com/tendermint/tm-db"
	"github.com/xlab/treeprint"
)

// RWTree provides an abstraction over IAVL that maintains separate read and write paths. Reads are routed to the most
// recently saved version of the tree - which provides immutable access. Writes are routed to a working tree that is
// mutable. On save the working tree is saved to DB, frozen, and replaces the previous immutable read tree.
type RWTree struct {
	// Working tree accumulating writes
	tree *MutableTree
	// Read-only tree serving previous state
	*ImmutableTree
	// Have any writes occurred since last save
	updated bool
}

// Creates a concurrency safe version of an IAVL tree whereby reads are routed to the last saved tree.
// Writes must be serialised (as they are within a commit for example).
func NewRWTree(db dbm.DB, cacheSize int) *RWTree {
	tree := NewMutableTree(db, cacheSize)
	return &RWTree{
		tree:          tree,
		ImmutableTree: &ImmutableTree{iavl.NewImmutableTree(db, cacheSize)},
	}
}

// Tries to load the execution state from DB, returns nil with no error if no state found
func (rwt *RWTree) Load(version int64, overwriting bool) error {
	const errHeader = "RWTree.Load():"
	if version <= 0 {
		return fmt.Errorf("%s trying to load from non-positive version %d", errHeader, version)
	}
	err := rwt.tree.Load(version, overwriting)
	if err != nil {
		return fmt.Errorf("%s loading version %d: %v", errHeader, version, err)
	}
	// Set readTree at commit point == tree
	rwt.ImmutableTree, err = rwt.tree.GetImmutable(version)
	if err != nil {
		return fmt.Errorf("%s loading version %d: %v", errHeader, version, err)
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
	rwt.updated = false
	return hash, version, nil
}

func (rwt *RWTree) Set(key, value []byte) bool {
	rwt.updated = true
	return rwt.tree.Set(key, value)
}

func (rwt *RWTree) Delete(key []byte) ([]byte, bool) {
	rwt.updated = true
	return rwt.tree.Remove(key)
}

// Returns true if there have been any writes since last save
func (rwt *RWTree) Updated() bool {
	return rwt.updated
}

func (rwt *RWTree) GetImmutable(version int64) (*ImmutableTree, error) {
	return rwt.tree.GetImmutable(version)
}

func (rwt *RWTree) IterateWriteTree(start, end []byte, ascending bool, fn func(key []byte, value []byte) error) error {
	return rwt.tree.IterateWriteTree(start, end, ascending, fn)
}

// Tree printing

func (rwt *RWTree) Dump() string {
	tree := treeprint.New()
	AddTreePrintTree("ReadTree", tree, rwt)
	AddTreePrintTree("WriteTree", tree, rwt.tree)
	return tree.String()
}

func AddTreePrintTree(edge string, tree treeprint.Tree, rwt KVCallbackIterableReader) {
	tree = tree.AddBranch(fmt.Sprintf("%q", edge))
	rwt.Iterate(nil, nil, true, func(key []byte, value []byte) error {
		tree.AddNode(fmt.Sprintf("%q -> %q", string(key), string(value)))
		return nil
	})
}
