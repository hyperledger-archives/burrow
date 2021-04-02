package storage

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/cosmos/iavl"
	dbm "github.com/tendermint/tm-db"
	"github.com/xlab/treeprint"
)

// RWTree provides an abstraction over IAVL that maintains separate read and write paths. Reads are routed to the most
// recently saved version of the tree - which provides immutable access. Writes are routed to a working tree that is
// mutable. On save the working tree is saved to DB, frozen, and replaces the previous immutable read tree.
type RWTree struct {
	// Synchronise 'write side access' - i.e. to the write tree and updated flag (which only read-locks this mutex)
	sync.RWMutex
	// Working tree accumulating writes
	tree *MutableTree
	// Read-only tree serving previous state
	readTree atomic.Value // *ImmutableTree
	// Have any writes occurred since last save
	updated bool
}

var _ KVCallbackIterableReader = &RWTree{}
var _ Versioned = &RWTree{}

// Creates a concurrency safe version of an IAVL tree whereby writes go a latest working tree and reads are routed to
// the last saved tree. All methods are safe for multiple readers and writers.
func NewRWTree(db dbm.DB, cacheSize int) (*RWTree, error) {
	tree, err := NewMutableTree(db, cacheSize)
	if err != nil {
		return nil, err
	}
	readTree := &ImmutableTree{iavl.NewImmutableTree(db, cacheSize)}
	rwt := &RWTree{
		tree: tree,
	}
	rwt.readTree.Store(readTree)
	return rwt, nil
}

// Write-side write methods - of the mutable tree - synchronised by write-lock of RWMutex

// Tries to load the execution state from DB, returns nil with no error if no state found
func (rwt *RWTree) Load(version int64, overwriting bool) error {
	rwt.Lock()
	defer rwt.Unlock()
	const errHeader = "RWTree.Load():"
	if version <= 0 {
		return fmt.Errorf("%s trying to load from non-positive version %d", errHeader, version)
	}
	err := rwt.tree.Load(version, overwriting)
	if err != nil {
		return fmt.Errorf("%s loading version %d: %v", errHeader, version, err)
	}
	// Set readTree at commit point == tree
	readTree, err := rwt.tree.GetImmutable(version)
	if err != nil {
		return fmt.Errorf("%s loading version %d: %v", errHeader, version, err)
	}
	rwt.readTree.Store(readTree)
	rwt.updated = false

	return nil
}

// Save the current write tree making writes accessible from read tree.
func (rwt *RWTree) Save() ([]byte, int64, error) {
	rwt.Lock()
	defer rwt.Unlock()
	// save state at a new version may still be orphaned before we save the version against the hash
	hash, version, err := rwt.tree.SaveVersion()
	if err != nil {
		return nil, 0, fmt.Errorf("could not save RWTree: %v", err)
	}
	// Take an immutable reference to the tree we just saved for querying
	readTree, err := rwt.tree.GetImmutable(version)
	if err != nil {
		return nil, 0, fmt.Errorf("RWTree.Save() could not obtain immutable read loadOrCreateTree: %v", err)
	}
	rwt.readTree.Store(readTree)
	rwt.updated = false
	return hash, version, nil
}

func (rwt *RWTree) Set(key, value []byte) bool {
	rwt.Lock()
	defer rwt.Unlock()
	rwt.updated = true
	return rwt.tree.Set(key, value)
}

func (rwt *RWTree) Delete(key []byte) ([]byte, bool) {
	rwt.Lock()
	defer rwt.Unlock()
	rwt.updated = true
	return rwt.tree.Remove(key)
}

// Write-side read methods - of the mutable tree - synchronised by read-lock of RWMutex

// Returns true if there have been any writes since last save
func (rwt *RWTree) Updated() bool {
	rwt.RLock()
	defer rwt.RUnlock()
	return rwt.updated
}

func (rwt *RWTree) GetImmutable(version int64) (*ImmutableTree, error) {
	rwt.RLock()
	defer rwt.RUnlock()
	return rwt.tree.GetImmutable(version)
}

func (rwt *RWTree) IterateWriteTree(start, end []byte, ascending bool, fn func(key []byte, value []byte) error) error {
	rwt.RLock()
	defer rwt.RUnlock()
	return rwt.tree.IterateWriteTree(start, end, ascending, fn)
}

// Read-side read methods - of the immutable read tree (previous saved state) - synchronised solely via atomic.Value

func (rwt *RWTree) Hash() []byte {
	return rwt.readTree.Load().(*ImmutableTree).Hash()
}

func (rwt *RWTree) Version() int64 {
	return rwt.readTree.Load().(*ImmutableTree).Version()
}

func (rwt *RWTree) Get(key []byte) ([]byte, error) {
	return rwt.readTree.Load().(*ImmutableTree).Get(key)
}

func (rwt *RWTree) Has(key []byte) (bool, error) {
	return rwt.readTree.Load().(*ImmutableTree).Has(key)
}

func (rwt *RWTree) Iterate(low, high []byte, ascending bool, fn func(key []byte, value []byte) error) error {
	return rwt.readTree.Load().(*ImmutableTree).Iterate(low, high, ascending, fn)
}

// Tree printing

func (rwt *RWTree) Dump() string {
	tree := treeprint.New()
	err := AddTreePrintTree("ReadTree", tree, rwt)
	if err != nil {
		return fmt.Sprintf("Error printing loadOrCreateTree: %v", err)
	}
	err = AddTreePrintTree("WriteTree", tree, rwt.tree)
	if err != nil {
		return fmt.Sprintf("Error printing loadOrCreateTree: %v", err)
	}
	return tree.String()
}

func AddTreePrintTree(edge string, tree treeprint.Tree, rwt KVCallbackIterableReader) error {
	tree = tree.AddBranch(fmt.Sprintf("%q", edge))
	return rwt.Iterate(nil, nil, true, func(key []byte, value []byte) error {
		tree.AddNode(fmt.Sprintf("%q -> %q", string(key), string(value)))
		return nil
	})
}
