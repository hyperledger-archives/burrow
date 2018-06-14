package iavl

// NOTE: This file favors int64 as opposed to int for size/counts.
// The Tree on the other hand favors int.  This is intentional.

import (
	"bytes"
	"fmt"
	"io"

	"github.com/tendermint/go-amino"
	"github.com/tendermint/iavl/sha256truncated"
)

// Node represents a node in a Tree.
type Node struct {
	key       []byte
	value     []byte
	version   int64
	height    int8
	size      int64
	hash      []byte
	leftHash  []byte
	leftNode  *Node
	rightHash []byte
	rightNode *Node
	persisted bool
}

// NewNode returns a new node from a key, value and version.
func NewNode(key []byte, value []byte, version int64) *Node {
	return &Node{
		key:     key,
		value:   value,
		height:  0,
		size:    1,
		version: version,
	}
}

// MakeNode constructs an *Node from an encoded byte slice.
//
// The new node doesn't have its hash saved or set.  The caller must set it
// afterwards.
func MakeNode(buf []byte) (node *Node, err error) {
	node = &Node{}

	// Keeps track of bytes read.
	n := 0

	// Read node header.
	node.height, n, err = amino.DecodeInt8(buf)
	if err != nil {
		return nil, err
	}
	buf = buf[n:]

	node.size, n, err = amino.DecodeInt64(buf)
	if err != nil {
		return nil, err
	}
	buf = buf[n:]

	node.version, n, err = amino.DecodeInt64(buf)
	if err != nil {
		return nil, err
	}
	buf = buf[n:]

	node.key, n, err = amino.DecodeByteSlice(buf)
	if err != nil {
		return nil, err
	}
	buf = buf[n:]

	// Read node body.

	if node.isLeaf() {
		node.value, _, err = amino.DecodeByteSlice(buf)
		if err != nil {
			return nil, err
		}
	} else { // Read children.
		leftHash, n, err := amino.DecodeByteSlice(buf)
		if err != nil {
			return nil, err
		}
		buf = buf[n:]

		rightHash, _, err := amino.DecodeByteSlice(buf)
		if err != nil {
			return nil, err
		}
		node.leftHash = leftHash
		node.rightHash = rightHash
	}
	return node, nil
}

// String returns a string representation of the node.
func (node *Node) String() string {
	if len(node.hash) == 0 {
		return "<no hash>"
	} else {
		return fmt.Sprintf("%x", node.hash)
	}
}

// clone creates a shallow copy of a node with its hash set to nil.
func (node *Node) clone(version int64) *Node {
	if node.isLeaf() {
		panic("Attempt to copy a leaf node")
	}
	return &Node{
		key:       node.key,
		height:    node.height,
		version:   version,
		size:      node.size,
		hash:      nil,
		leftHash:  node.leftHash,
		leftNode:  node.leftNode,
		rightHash: node.rightHash,
		rightNode: node.rightNode,
		persisted: false,
	}
}

func (node *Node) isLeaf() bool {
	return node.height == 0
}

// Check if the node has a descendant with the given key.
func (node *Node) has(t *Tree, key []byte) (has bool) {
	if bytes.Equal(node.key, key) {
		return true
	}
	if node.isLeaf() {
		return false
	}
	if bytes.Compare(key, node.key) < 0 {
		return node.getLeftNode(t).has(t, key)
	} else {
		return node.getRightNode(t).has(t, key)
	}
}

// Get a key under the node.
func (node *Node) get(t *Tree, key []byte) (index int64, value []byte) {
	if node.isLeaf() {
		switch bytes.Compare(node.key, key) {
		case -1:
			return 1, nil
		case 1:
			return 0, nil
		default:
			return 0, node.value
		}
	}

	if bytes.Compare(key, node.key) < 0 {
		return node.getLeftNode(t).get(t, key)
	} else {
		rightNode := node.getRightNode(t)
		index, value = rightNode.get(t, key)
		index += node.size - rightNode.size
		return index, value
	}
}

func (node *Node) getByIndex(t *Tree, index int64) (key []byte, value []byte) {
	if node.isLeaf() {
		if index == 0 {
			return node.key, node.value
		} else {
			return nil, nil
		}
	} else {
		// TODO: could improve this by storing the
		// sizes as well as left/right hash.
		leftNode := node.getLeftNode(t)
		if index < leftNode.size {
			return leftNode.getByIndex(t, index)
		} else {
			return node.getRightNode(t).getByIndex(t, index-leftNode.size)
		}
	}
}

// Computes the hash of the node without computing its descendants. Must be
// called on nodes which have descendant node hashes already computed.
func (node *Node) _hash() []byte {
	if node.hash != nil {
		return node.hash
	}

	h := sha256truncated.New()
	buf := new(bytes.Buffer)
	if err := node.writeHashBytes(buf); err != nil {
		panic(err)
	}
	h.Write(buf.Bytes())
	node.hash = h.Sum(nil)

	return node.hash
}

// Hash the node and its descendants recursively. This usually mutates all
// descendant nodes. Returns the node hash and number of nodes hashed.
func (node *Node) hashWithCount() ([]byte, int64) {
	if node.hash != nil {
		return node.hash, 0
	}

	h := sha256truncated.New()
	buf := new(bytes.Buffer)
	hashCount, err := node.writeHashBytesRecursively(buf)
	if err != nil {
		panic(err)
	}
	h.Write(buf.Bytes())
	node.hash = h.Sum(nil)

	return node.hash, hashCount + 1
}

// Writes the node's hash to the given io.Writer. This function expects
// child hashes to be already set.
func (node *Node) writeHashBytes(w io.Writer) (err error) {
	err = amino.EncodeInt8(w, node.height)
	if err == nil {
		err = amino.EncodeInt64(w, node.size)
	}
	if err == nil {
		err = amino.EncodeInt64(w, node.version)
	}

	// Key is not written for inner nodes, unlike writeBytes.

	if node.isLeaf() {
		if err == nil {
			err = amino.EncodeByteSlice(w, node.key)
		}
		if err == nil {
			err = amino.EncodeByteSlice(w, node.value)
		}
	} else {
		if node.leftHash == nil || node.rightHash == nil {
			panic("Found an empty child hash")
		}
		if err == nil {
			err = amino.EncodeByteSlice(w, node.leftHash)
		}
		if err == nil {
			err = amino.EncodeByteSlice(w, node.rightHash)
		}
	}
	return
}

// Writes the node's hash to the given io.Writer.
// This function has the side-effect of calling hashWithCount.
func (node *Node) writeHashBytesRecursively(w io.Writer) (hashCount int64, err error) {
	if node.leftNode != nil {
		leftHash, leftCount := node.leftNode.hashWithCount()
		node.leftHash = leftHash
		hashCount += leftCount
	}
	if node.rightNode != nil {
		rightHash, rightCount := node.rightNode.hashWithCount()
		node.rightHash = rightHash
		hashCount += rightCount
	}
	err = node.writeHashBytes(w)

	return
}

// Writes the node as a serialized byte slice to the supplied io.Writer.
func (node *Node) writeBytes(w io.Writer) (err error) {
	err = amino.EncodeInt8(w, node.height)
	if err == nil {
		err = amino.EncodeInt64(w, node.size)
	}
	if err == nil {
		err = amino.EncodeInt64(w, node.version)
	}

	// Unlike writeHashBytes, key is written for inner nodes.
	if err == nil {
		err = amino.EncodeByteSlice(w, node.key)
	}

	if node.isLeaf() {
		if err == nil {
			err = amino.EncodeByteSlice(w, node.value)
		}
	} else {
		if node.leftHash == nil {
			panic("node.leftHash was nil in writeBytes")
		}
		if err == nil {
			err = amino.EncodeByteSlice(w, node.leftHash)
		}

		if node.rightHash == nil {
			panic("node.rightHash was nil in writeBytes")
		}
		if err == nil {
			err = amino.EncodeByteSlice(w, node.rightHash)
		}
	}
	return
}

func (node *Node) set(t *Tree, key []byte, value []byte) (
	newSelf *Node, updated bool, orphaned []*Node,
) {
	version := t.version + 1

	if node.isLeaf() {
		switch bytes.Compare(key, node.key) {
		case -1:
			return &Node{
				key:       node.key,
				height:    1,
				size:      2,
				leftNode:  NewNode(key, value, version),
				rightNode: node,
				version:   version,
			}, false, []*Node{}
		case 1:
			return &Node{
				key:       key,
				height:    1,
				size:      2,
				leftNode:  node,
				rightNode: NewNode(key, value, version),
				version:   version,
			}, false, []*Node{}
		default:
			return NewNode(key, value, version), true, []*Node{node}
		}
	} else {
		orphaned = append(orphaned, node)
		node = node.clone(version)

		if bytes.Compare(key, node.key) < 0 {
			var leftOrphaned []*Node
			node.leftNode, updated, leftOrphaned = node.getLeftNode(t).set(t, key, value)
			node.leftHash = nil // leftHash is yet unknown
			orphaned = append(orphaned, leftOrphaned...)
		} else {
			var rightOrphaned []*Node
			node.rightNode, updated, rightOrphaned = node.getRightNode(t).set(t, key, value)
			node.rightHash = nil // rightHash is yet unknown
			orphaned = append(orphaned, rightOrphaned...)
		}

		if updated {
			return node, updated, orphaned
		} else {
			node.calcHeightAndSize(t)
			newNode, balanceOrphaned := node.balance(t)
			return newNode, updated, append(orphaned, balanceOrphaned...)
		}
	}
}

// newHash/newNode: The new hash or node to replace node after remove.
// newKey: new leftmost leaf key for tree after successfully removing 'key' if changed.
// value: removed value.
func (node *Node) remove(t *Tree, key []byte) (
	newHash []byte, newNode *Node, newKey []byte, value []byte, orphaned []*Node,
) {
	version := t.version + 1

	if node.isLeaf() {
		if bytes.Equal(key, node.key) {
			return nil, nil, nil, node.value, []*Node{node}
		}
		return node.hash, node, nil, nil, orphaned
	}

	if bytes.Compare(key, node.key) < 0 {
		var newLeftHash []byte
		var newLeftNode *Node

		newLeftHash, newLeftNode, newKey, value, orphaned =
			node.getLeftNode(t).remove(t, key)

		if len(orphaned) == 0 {
			return node.hash, node, nil, value, orphaned
		} else if newLeftHash == nil && newLeftNode == nil { // left node held value, was removed
			return node.rightHash, node.rightNode, node.key, value, orphaned
		}
		orphaned = append(orphaned, node)

		newNode := node.clone(version)
		newNode.leftHash, newNode.leftNode = newLeftHash, newLeftNode
		newNode.calcHeightAndSize(t)
		newNode, balanceOrphaned := newNode.balance(t)

		return newNode.hash, newNode, newKey, value, append(orphaned, balanceOrphaned...)
	} else {
		var newRightHash []byte
		var newRightNode *Node

		newRightHash, newRightNode, newKey, value, orphaned =
			node.getRightNode(t).remove(t, key)

		if len(orphaned) == 0 {
			return node.hash, node, nil, value, orphaned
		} else if newRightHash == nil && newRightNode == nil { // right node held value, was removed
			return node.leftHash, node.leftNode, nil, value, orphaned
		}
		orphaned = append(orphaned, node)

		newNode := node.clone(version)
		newNode.rightHash, newNode.rightNode = newRightHash, newRightNode
		if newKey != nil {
			newNode.key = newKey
		}
		newNode.calcHeightAndSize(t)
		newNode, balanceOrphaned := newNode.balance(t)

		return newNode.hash, newNode, nil, value, append(orphaned, balanceOrphaned...)
	}
}

func (node *Node) getLeftNode(t *Tree) *Node {
	if node.leftNode != nil {
		return node.leftNode
	}
	return t.ndb.GetNode(node.leftHash)
}

func (node *Node) getRightNode(t *Tree) *Node {
	if node.rightNode != nil {
		return node.rightNode
	}
	return t.ndb.GetNode(node.rightHash)
}

// Rotate right and return the new node and orphan.
func (node *Node) rotateRight(t *Tree) (newNode *Node, orphan *Node) {
	version := t.version + 1

	// TODO: optimize balance & rotate.
	node = node.clone(version)
	l := node.getLeftNode(t)
	_l := l.clone(version)

	_lrHash, _lrCached := _l.rightHash, _l.rightNode
	_l.rightHash, _l.rightNode = node.hash, node
	node.leftHash, node.leftNode = _lrHash, _lrCached

	node.calcHeightAndSize(t)
	_l.calcHeightAndSize(t)

	return _l, l
}

// Rotate left and return the new node and orphan.
func (node *Node) rotateLeft(t *Tree) (newNode *Node, orphan *Node) {
	version := t.version + 1

	// TODO: optimize balance & rotate.
	node = node.clone(version)
	r := node.getRightNode(t)
	_r := r.clone(version)

	_rlHash, _rlCached := _r.leftHash, _r.leftNode
	_r.leftHash, _r.leftNode = node.hash, node
	node.rightHash, node.rightNode = _rlHash, _rlCached

	node.calcHeightAndSize(t)
	_r.calcHeightAndSize(t)

	return _r, r
}

// NOTE: mutates height and size
func (node *Node) calcHeightAndSize(t *Tree) {
	node.height = maxInt8(node.getLeftNode(t).height, node.getRightNode(t).height) + 1
	node.size = node.getLeftNode(t).size + node.getRightNode(t).size
}

func (node *Node) calcBalance(t *Tree) int {
	return int(node.getLeftNode(t).height) - int(node.getRightNode(t).height)
}

// NOTE: assumes that node can be modified
// TODO: optimize balance & rotate
func (node *Node) balance(t *Tree) (newSelf *Node, orphaned []*Node) {
	if node.persisted {
		panic("Unexpected balance() call on persisted node")
	}
	balance := node.calcBalance(t)

	if balance > 1 {
		if node.getLeftNode(t).calcBalance(t) >= 0 {
			// Left Left Case
			newNode, orphaned := node.rotateRight(t)
			return newNode, []*Node{orphaned}
		} else {
			// Left Right Case
			var leftOrphaned *Node

			left := node.getLeftNode(t)
			node.leftHash = nil
			node.leftNode, leftOrphaned = left.rotateLeft(t)
			newNode, rightOrphaned := node.rotateRight(t)

			return newNode, []*Node{left, leftOrphaned, rightOrphaned}
		}
	}
	if balance < -1 {
		if node.getRightNode(t).calcBalance(t) <= 0 {
			// Right Right Case
			newNode, orphaned := node.rotateLeft(t)
			return newNode, []*Node{orphaned}
		} else {
			// Right Left Case
			var rightOrphaned *Node

			right := node.getRightNode(t)
			node.rightHash = nil
			node.rightNode, rightOrphaned = right.rotateRight(t)
			newNode, leftOrphaned := node.rotateLeft(t)

			return newNode, []*Node{right, leftOrphaned, rightOrphaned}
		}
	}
	// Nothing changed
	return node, []*Node{}
}

// traverse is a wrapper over traverseInRange when we want the whole tree
func (node *Node) traverse(t *Tree, ascending bool, cb func(*Node) bool) bool {
	return node.traverseInRange(t, nil, nil, ascending, false, 0, func(node *Node, depth uint8) bool {
		return cb(node)
	})
}

func (node *Node) traverseWithDepth(t *Tree, ascending bool, cb func(*Node, uint8) bool) bool {
	return node.traverseInRange(t, nil, nil, ascending, false, 0, cb)
}

func (node *Node) traverseInRange(t *Tree, start, end []byte, ascending bool, inclusive bool, depth uint8, cb func(*Node, uint8) bool) bool {
	afterStart := start == nil || bytes.Compare(start, node.key) <= 0
	beforeEnd := end == nil || bytes.Compare(node.key, end) < 0
	if inclusive {
		beforeEnd = end == nil || bytes.Compare(node.key, end) <= 0
	}

	stop := false
	if afterStart && beforeEnd {
		// IterateRange ignores this if not leaf
		stop = cb(node, depth)
	}
	if stop {
		return stop
	}
	if node.isLeaf() {
		return stop
	}

	if ascending {
		// check lower nodes, then higher
		if afterStart {
			stop = node.getLeftNode(t).traverseInRange(t, start, end, ascending, inclusive, depth+1, cb)
		}
		if stop {
			return stop
		}
		if beforeEnd {
			stop = node.getRightNode(t).traverseInRange(t, start, end, ascending, inclusive, depth+1, cb)
		}
	} else {
		// check the higher nodes first
		if beforeEnd {
			stop = node.getRightNode(t).traverseInRange(t, start, end, ascending, inclusive, depth+1, cb)
		}
		if stop {
			return stop
		}
		if afterStart {
			stop = node.getLeftNode(t).traverseInRange(t, start, end, ascending, inclusive, depth+1, cb)
		}
	}

	return stop
}

// Only used in testing...
func (node *Node) lmd(t *Tree) *Node {
	if node.isLeaf() {
		return node
	}
	return node.getLeftNode(t).lmd(t)
}
