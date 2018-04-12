package iavl

import (
	"fmt"
)

// orphaningTree is a tree which keeps track of orphaned nodes.
type orphaningTree struct {
	*Tree

	// A map of orphan hash to orphan version.
	// The version stored here is the one at which the orphan's lifetime
	// begins.
	orphans map[string]int64
}

// newOrphaningTree creates a new orphaning tree from the given *Tree.
func newOrphaningTree(t *Tree) *orphaningTree {
	return &orphaningTree{
		Tree:    t,
		orphans: map[string]int64{},
	}
}

// Set a key on the underlying tree while storing the orphaned nodes.
func (tree *orphaningTree) Set(key, value []byte) bool {
	orphaned, updated := tree.Tree.set(key, value)
	tree.addOrphans(orphaned)
	return updated
}

// Remove a key from the underlying tree while storing the orphaned nodes.
func (tree *orphaningTree) Remove(key []byte) ([]byte, bool) {
	val, orphaned, removed := tree.Tree.remove(key)
	tree.addOrphans(orphaned)
	return val, removed
}

// SaveAs saves the underlying Tree and assigns it a new version.
// Saves orphans too.
func (tree *orphaningTree) SaveAs(version int64) {
	if version != tree.version+1 {
		panic(fmt.Sprintf("Expected to save version %d but tried to save %d", tree.version+1, version))
	}
	if tree.root == nil {
		// There can still be orphans, for example if the root is the node being
		// removed.
		tree.ndb.SaveOrphans(version, tree.orphans)
		tree.ndb.SaveEmptyRoot(version)
	} else {
		debug("SAVE TREE %v\n", version)
		// Save the current tree.
		tree.ndb.SaveBranch(tree.root)
		tree.ndb.SaveOrphans(version, tree.orphans)
		tree.ndb.SaveRoot(tree.root, version)
	}
	tree.ndb.Commit()
	tree.version = version
}

// Add orphans to the orphan list. Doesn't write to disk.
func (tree *orphaningTree) addOrphans(orphans []*Node) {
	for _, node := range orphans {
		if !node.persisted {
			// We don't need to orphan nodes that were never persisted.
			continue
		}
		if len(node.hash) == 0 {
			panic("Expected to find node hash, but was empty")
		}
		tree.orphans[string(node.hash)] = node.version
	}
}
