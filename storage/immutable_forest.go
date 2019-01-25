package storage

import (
	"fmt"

	"github.com/xlab/treeprint"

	lru "github.com/hashicorp/golang-lru"
	dbm "github.com/tendermint/tendermint/libs/db"
)

type ImmutableForest struct {
	// Store of tree prefix -> last commitID (version + hash) - serves as a set of all known trees and provides a global hash
	commitsTree KVCallbackIterableReader
	treeDB      dbm.DB
	// Cache for frequently used trees
	treeCache *lru.Cache
	// Cache size is used in multiple places - for the LRU cache and node cache for any trees created - it probably
	// makes sense for them to be roughly the same size
	cacheSize int
	// Determines whether we use LoadVersionForOverwriting on underlying MutableTrees - since ImmutableForest is used
	// by MutableForest in a writing context sometimes we do need to load a version destructively
	overwriting bool
}

type ForestOption func(*ImmutableForest)

var WithOverwriting ForestOption = func(imf *ImmutableForest) { imf.overwriting = true }

func NewImmutableForest(commitsTree KVCallbackIterableReader, treeDB dbm.DB, cacheSize int,
	options ...ForestOption) (*ImmutableForest, error) {
	cache, err := lru.New(cacheSize)
	if err != nil {
		return nil, fmt.Errorf("NewImmutableForest() could not create cache: %v", err)
	}
	imf := &ImmutableForest{
		commitsTree: commitsTree,
		treeDB:      treeDB,
		treeCache:   cache,
		cacheSize:   cacheSize,
	}
	for _, opt := range options {
		opt(imf)
	}
	return imf, nil
}

func (imf *ImmutableForest) Iterate(start, end []byte, ascending bool, fn func(prefix []byte, tree KVCallbackIterableReader) error) error {
	return imf.commitsTree.Iterate(start, end, ascending, func(prefix []byte, _ []byte) error {
		rwt, err := imf.tree(prefix)
		if err != nil {
			return err
		}
		return fn(prefix, rwt)
	})
}

func (imf *ImmutableForest) IterateRWTree(start, end []byte, ascending bool, fn func(prefix []byte, tree *RWTree) error) error {
	return imf.commitsTree.Iterate(start, end, ascending, func(prefix []byte, _ []byte) error {
		rwt, err := imf.tree(prefix)
		if err != nil {
			return err
		}
		return fn(prefix, rwt)
	})
}

// Get the tree at prefix for making reads
func (imf *ImmutableForest) Reader(prefix []byte) (KVCallbackIterableReader, error) {
	return imf.tree(prefix)
}

// Lazy load tree
func (imf *ImmutableForest) tree(prefix []byte) (*RWTree, error) {
	// Try cache
	if value, ok := imf.treeCache.Get(string(prefix)); ok {
		return value.(*RWTree), nil
	}
	// Not in caches but non-negative version - we should be able to load into memory
	return imf.loadOrCreateTree(prefix)
}

func (imf *ImmutableForest) commitID(prefix []byte) (*CommitID, error) {
	bs := imf.commitsTree.Get(prefix)
	if bs == nil {
		return new(CommitID), nil
	}
	commitID, err := UnmarshalCommitID(bs)
	if err != nil {
		return nil, fmt.Errorf("could not get commitID for prefix %X: %v", prefix, err)
	}
	return commitID, nil
}

func (imf *ImmutableForest) loadOrCreateTree(prefix []byte) (*RWTree, error) {
	const errHeader = "ImmutableForest.loadOrCreateTree():"
	tree := imf.newTree(prefix)
	commitID, err := imf.commitID(prefix)
	if err != nil {
		return nil, fmt.Errorf("%s %v", errHeader, err)
	}
	if commitID.Version == 0 {
		// This is the first time we have been asked to load this tree
		return imf.newTree(prefix), nil
	}
	err = tree.Load(commitID.Version, imf.overwriting)
	if err != nil {
		return nil, fmt.Errorf("%s could not load tree: %v", errHeader, err)
	}
	return tree, nil
}

// Create a new in-memory IAVL tree
func (imf *ImmutableForest) newTree(prefix []byte) *RWTree {
	p := string(prefix)
	tree := NewRWTree(NewPrefixDB(imf.treeDB, p), imf.cacheSize)
	imf.treeCache.Add(p, tree)
	return tree
}

func (imf *ImmutableForest) Dump() string {
	dump := treeprint.New()
	imf.Iterate(nil, nil, true, func(prefix []byte, tree KVCallbackIterableReader) error {
		AddTreePrintTree(string(prefix), dump, tree)
		return nil
	})
	return dump.String()
}
