package iavl

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"
	dbm "github.com/tendermint/tmlibs/db"
)

var ErrVersionDoesNotExist = fmt.Errorf("version does not exist")

// VersionedTree is a persistent tree which keeps track of versions.
type VersionedTree struct {
	*orphaningTree                 // The current, working tree.
	versions       map[int64]*Tree // The previous, saved versions of the tree.
	ndb            *nodeDB
}

// NewVersionedTree returns a new tree with the specified cache size and datastore.
func NewVersionedTree(db dbm.DB, cacheSize int) *VersionedTree {
	ndb := newNodeDB(db, cacheSize)
	head := &Tree{ndb: ndb}

	return &VersionedTree{
		orphaningTree: newOrphaningTree(head),
		versions:      map[int64]*Tree{},
		ndb:           ndb,
	}
}

// IsEmpty returns whether or not the tree has any keys. Only trees that are
// not empty can be saved.
func (tree *VersionedTree) IsEmpty() bool {
	return tree.orphaningTree.Size() == 0
}

// VersionExists returns whether or not a version exists.
func (tree *VersionedTree) VersionExists(version int64) bool {
	_, ok := tree.versions[version]
	return ok
}

// Tree returns the current working tree.
func (tree *VersionedTree) Tree() *Tree {
	return tree.orphaningTree.Tree
}

// Hash returns the hash of the latest saved version of the tree, as returned
// by SaveVersion. If no versions have been saved, Hash returns nil.
func (tree *VersionedTree) Hash() []byte {
	if tree.version > 0 {
		return tree.versions[tree.version].Hash()
	}
	return nil
}

// String returns a string representation of the tree.
func (tree *VersionedTree) String() string {
	return tree.ndb.String()
}

// Set sets a key in the working tree. Nil values are not supported.
func (tree *VersionedTree) Set(key, val []byte) bool {
	return tree.orphaningTree.Set(key, val)
}

// Remove removes a key from the working tree.
func (tree *VersionedTree) Remove(key []byte) ([]byte, bool) {
	return tree.orphaningTree.Remove(key)
}

// Load the latest versioned tree from disk.
//
// Returns the version number of the latest version found
func (tree *VersionedTree) Load() (int64, error) {
	return tree.LoadVersion(0)
}

// Load a versioned tree from disk.
//
// If version is 0, the latest version is loaded.
//
// Returns the version number of the latest version found
func (tree *VersionedTree) LoadVersion(targetVersion int64) (int64, error) {
	roots, err := tree.ndb.getRoots()
	if err != nil {
		return 0, err
	}
	if len(roots) == 0 {
		return 0, nil
	}

	// Load all roots from the database.
	latestVersion := int64(0)
	for version, root := range roots {

		// Construct a tree manually.
		t := &Tree{}
		t.ndb = tree.ndb
		t.version = version
		if len(root) != 0 {
			t.root = tree.ndb.GetNode(root)
		}
		tree.versions[version] = t

		if version > latestVersion &&
			(targetVersion == 0 || version <= targetVersion) {

			latestVersion = version
		}
	}

	// Validate latestVersion
	if !(targetVersion == 0 || latestVersion == targetVersion) {
		return latestVersion, fmt.Errorf("Wanted to load target %v but only found up to %v",
			targetVersion, latestVersion)
	}

	// Set the working tree to a copy of the latest.
	tree.orphaningTree = newOrphaningTree(
		tree.versions[latestVersion].clone(),
	)

	return latestVersion, nil
}

// Rollback resets the working tree to the latest saved version, discarding
// any unsaved modifications.
func (tree *VersionedTree) Rollback() {
	if tree.version > 0 {
		tree.orphaningTree = newOrphaningTree(
			tree.versions[tree.version].clone(),
		)
	} else {
		tree.orphaningTree = newOrphaningTree(&Tree{ndb: tree.ndb, version: 0})
	}
}

// GetVersioned gets the value at the specified key and version.
func (tree *VersionedTree) GetVersioned(key []byte, version int64) (
	index int, value []byte,
) {
	if t, ok := tree.versions[version]; ok {
		return t.Get(key)
	}
	return -1, nil
}

// SaveVersion saves a new tree version to disk, based on the current state of
// the tree. Returns the hash and new version number.
func (tree *VersionedTree) SaveVersion() ([]byte, int64, error) {
	version := tree.version + 1

	if _, ok := tree.versions[version]; ok {
		// Same hash means idempotent.  Return success.
		var existingHash = tree.versions[version].Hash()
		var newHash = tree.orphaningTree.Hash()
		if bytes.Equal(existingHash, newHash) {
			tree.orphaningTree = newOrphaningTree(tree.versions[version].clone())
			return existingHash, version, nil
		}
		return nil, version, errors.Errorf("version %d was already saved to different hash %X (existing hash %X)",
			version, newHash, existingHash)
	}

	// Persist version and stash to .versions.
	tree.orphaningTree.SaveAs(version)
	tree.versions[version] = tree.orphaningTree.Tree

	// Set new working tree.
	tree.orphaningTree = newOrphaningTree(tree.orphaningTree.clone())

	return tree.Hash(), version, nil
}

// DeleteVersion deletes a tree version from disk. The version can then no
// longer be accessed.
func (tree *VersionedTree) DeleteVersion(version int64) error {
	if version == 0 {
		return errors.New("version must be greater than 0")
	}
	if version == tree.version {
		return errors.Errorf("cannot delete latest saved version (%d)", version)
	}
	if _, ok := tree.versions[version]; !ok {
		return errors.WithStack(ErrVersionDoesNotExist)
	}

	tree.ndb.DeleteVersion(version)
	tree.ndb.Commit()

	delete(tree.versions, version)

	return nil
}

// GetVersionedWithProof gets the value under the key at the specified version
// if it exists, or returns nil.  A proof of existence or absence is returned
// alongside the value.
func (tree *VersionedTree) GetVersionedWithProof(key []byte, version int64) ([]byte, KeyProof, error) {
	if t, ok := tree.versions[version]; ok {
		return t.GetWithProof(key)
	}
	return nil, nil, errors.WithStack(ErrVersionDoesNotExist)
}

// GetVersionedRangeWithProof gets key/value pairs within the specified range
// and limit. To specify a descending range, swap the start and end keys.
//
// Returns a list of keys, a list of values and a proof.
func (tree *VersionedTree) GetVersionedRangeWithProof(startKey, endKey []byte, limit int, version int64) ([][]byte, [][]byte, *KeyRangeProof, error) {
	if t, ok := tree.versions[version]; ok {
		return t.GetRangeWithProof(startKey, endKey, limit)
	}
	return nil, nil, nil, errors.WithStack(ErrVersionDoesNotExist)
}

// GetVersionedFirstInRangeWithProof gets the first key/value pair in the
// specified range, with a proof.
func (tree *VersionedTree) GetVersionedFirstInRangeWithProof(startKey, endKey []byte, version int64) ([]byte, []byte, *KeyFirstInRangeProof, error) {
	if t, ok := tree.versions[version]; ok {
		return t.GetFirstInRangeWithProof(startKey, endKey)
	}
	return nil, nil, nil, errors.WithStack(ErrVersionDoesNotExist)
}

// GetVersionedLastInRangeWithProof gets the last key/value pair in the
// specified range, with a proof.
func (tree *VersionedTree) GetVersionedLastInRangeWithProof(startKey, endKey []byte, version int64) ([]byte, []byte, *KeyLastInRangeProof, error) {
	if t, ok := tree.versions[version]; ok {
		return t.GetLastInRangeWithProof(startKey, endKey)
	}
	return nil, nil, nil, errors.WithStack(ErrVersionDoesNotExist)
}
