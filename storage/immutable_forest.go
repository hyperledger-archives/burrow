package storage

import (
	"fmt"

	"github.com/xlab/treeprint"

	lru "github.com/hashicorp/golang-lru"
	amino "github.com/tendermint/go-amino"
	dbm "github.com/tendermint/tendermint/libs/db"
)

type ImmutableForest struct {
	// Store of tree prefix -> last commitID (version + hash) - serves as a set of all known trees and provides a global hash
	commitsTree KVCallbackIterableReader
	treeDB      dbm.DB
	// Map of prefix -> tree for trees requiring a save
	dirty map[string]*RWTree
	// Cache for frequently used trees
	treeCache *lru.Cache
	cacheSize int
	codec     *amino.Codec
}

func NewImmutableForest(commitsTree KVCallbackIterableReader, treeDB dbm.DB, cacheSize int) (*ImmutableForest, error) {
	cache, err := lru.New(cacheSize)
	if err != nil {
		return nil, fmt.Errorf("NewRWForest() could not create cache: %v", err)
	}
	return &ImmutableForest{
		commitsTree: commitsTree,
		treeDB:      treeDB,
		dirty:       make(map[string]*RWTree),
		treeCache:   cache,
		cacheSize:   cacheSize,
		codec:       amino.NewCodec(),
	}, nil
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
	// Try dirty
	if tree, ok := imf.dirty[string(prefix)]; ok {
		return tree, nil
	}
	// If we don't have tree in either of our caches and we are on zeroth version this is the first time it has been
	// requested so create a new one
	commitID, err := imf.commitID(prefix)
	if err != nil {
		return nil, err
	}
	if commitID.Version == 0 {
		return imf.newTree(prefix), nil
	}
	// Not in caches but non-negative version - we should be able to load into memory
	return imf.loadTree(prefix)
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

func (imf *ImmutableForest) loadTree(prefix []byte) (*RWTree, error) {
	const errHeader = "RWForest.loadTree():"
	tree := imf.newTree(prefix)
	commitID, err := imf.commitID(prefix)
	if err != nil {
		return nil, fmt.Errorf("%s %v", errHeader, err)
	}
	err = tree.Load(commitID.Version)
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
