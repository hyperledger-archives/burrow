// Implementation of a trie based fairly loosely on a qp-trie (https://dotat.at/prog/qp/README.html) but for a branching
// factor of 256 + 1 (for terminal)

package trie

import (
	"bytes"
	"fmt"

	"github.com/xlab/treeprint"
)

type Address interface{}

type Trie struct {
	*Node
}

func NewTrie() *Trie {
	return &Trie{NewBranch(TerminalChar)}
}
func (trie *Trie) Get(key []byte) (value interface{}, exists bool) {
	_, _, n := trie.Descend(key)
	if n.IsLeaf() && bytes.Equal(key, n.key) {
		return n.value, true
	}
	return nil, false
}

func (trie *Trie) Set(key []byte, value interface{}) (updated bool) {
	// Insert proceeds in two stages - first we check to see if there is a slot in the existing tree structure in to
	// store our key-value pair. If there is not then we work out where we need to extend the trie.

	// Traverse as far as we can in order to find a possible matching key, otherwise the 'nearest' matching key
	// that matches as much as possible along (the indexed bytes) of a prefix of the key
	branch, _, child := trie.Descend(key)
	if child == nil {
		// ** Simple Insert **
		// No match for leaf and we have gone as deep as we can. Since child is nil there is a space in this branch for
		// the char of this key at branch's critical index so we can just populate that slot
		addLeaf(branch, key, value)
		return false
	}
	// Child is a exact match leaf
	if child.IsLeaf() && bytes.Equal(key, child.key) {
		// ** Update **
		// exact match - set
		child.value = value
		return true
	}

	// ** Split Insert **
	// Simple Insert and Update have failed so key and child.key must not match on some non-indexed key in the trie
	// We must introduce a new branch indexed on the critical char

	// If child is still a branch descend down the left-hand side to find a suitable 'nearest' leaf node
	// (sharing as much prefix as possible)
	for !child.IsLeaf() {
		branch = child
		// Assertion: A branch should always have at least one child and the should terminate with a leaf
		child = branch.children[0]
	}

	// Get divergent char index from our child leaf for which we know matches as much as possible along the path of
	// existing branch nodes
	nearestKey := child.key
	criticalIndex := criticalIndex(key, nearestKey)
	criticalChar := charAt(nearestKey, criticalIndex)
	// Rewind to root and descend to critical index
	branch, char, child := trie.Descend(key[:criticalIndex+1])
	if child == nil {
		panic("no child found traversing to critical index but if that was the case could have performed simple insert")
	}

	// Replace child at char of current branch with a new branch named twig
	// Create new branch to hold the current child at char
	twig := NewBranch(criticalIndex)
	// Add the new value to this branch
	addLeaf(twig, key, value)
	// Add the existing child to this branch as a sibling (they differ at the critical index)
	twig.Add(criticalChar, child)
	// Replace the child at char with the new branch twig
	branch.Remove(char)
	branch.Add(char, twig)
	return false
}

func (trie *Trie) Delete(key []byte) (deleted bool) {
	b, _, n := trie.Descend(key)
	if !n.IsLeaf() || !bytes.Equal(key, n.key) {
		// Key does not exist
		return false
	}
	// n is an exact match
	b.Remove(charAt(key, b.index))
	// n is orphaned
	// See if we need to contract the branch
	if len(b.children) == 1 {
		child := b.children[0]
		// Overwrite the contents of b with that of its first and only child - in effect it becomes its child and we have
		// eliminated the unnecessary branch
		b.Leaf = child.Leaf
		b.Branch = child.Branch
	}
	return true
}

func addLeaf(tn *Node, key []byte, value interface{}) {
	tn.Add(charAt(key, tn.index), &Node{Leaf: &Leaf{key: key, value: value}})
}

func (trie *Trie) Dump() string {
	tree := treeprint.New()
	buildTree("", tree, trie.Node)
	return tree.String()
}

func buildTree(edge string, tree treeprint.Tree, node *Node) {
	if node.IsBranch() {
		chars := node.childChars()
		if edge != "" {
			tree = tree.AddBranch(edge)
		}
		for i, child := range node.children {
			buildTree(fmt.Sprintf("'%s' @ %d", stringIndexOf(chars[i]), node.index), tree, child)
		}
	} else {
		tree.AddNode(fmt.Sprintf("%s -> %s: %v", edge, string(node.key), node.value))
	}
}
