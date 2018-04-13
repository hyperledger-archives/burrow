package iavl

import (
	"sort"

	"github.com/pkg/errors"
	cmn "github.com/tendermint/tmlibs/common"
)

// Chunk is a list of ordered nodes.
// It can be sorted, merged, exported from a tree and
// used to generate a new tree.
type Chunk []OrderedNodeData

// OrderedNodeData is the data to recreate a leaf node,
// along with a SortOrder to define a BFS insertion order.
type OrderedNodeData struct {
	SortOrder uint64
	NodeData
}

// NewOrderedNode creates the data from a leaf node.
func NewOrderedNode(leaf *Node, prefix uint64) OrderedNodeData {
	return OrderedNodeData{
		SortOrder: prefix,
		NodeData: NodeData{
			Key:   leaf.key,
			Value: leaf.value,
		},
	}
}

// getChunkHashes returns all the "checksum" hashes for
// the chunks that will be sent.
func getChunkHashes(tree *Tree, depth uint) ([][]byte, [][]byte, uint, error) {
	maxDepth := uint(tree.root.height / 2)
	if depth > maxDepth {
		return nil, nil, 0, errors.New("depth exceeds maximum allowed")
	}

	nodes := getNodes(tree, depth)
	hashes := make([][]byte, len(nodes))
	keys := make([][]byte, len(nodes))
	for i, n := range nodes {
		hashes[i] = n.hash
		keys[i] = n.key
	}
	return hashes, keys, depth, nil
}

// GetChunkHashesWithProofs takes a tree and returns the list of chunks with
// proofs that can be used to synchronize a tree across the network.
func GetChunkHashesWithProofs(tree *Tree) ([][]byte, []*InnerKeyProof, uint) {
	hashes, keys, depth, err := getChunkHashes(tree, uint(tree.root.height/2))
	if err != nil {
		cmn.PanicSanity(cmn.Fmt("GetChunkHashes: %s", err))
	}
	proofs := make([]*InnerKeyProof, len(keys))

	for i, k := range keys {
		proof, err := tree.getInnerWithProof(k)
		if err != nil {
			cmn.PanicSanity(cmn.Fmt("Error getting inner key proof: %s", err))
		}
		proofs[i] = proof
	}
	return hashes, proofs, depth
}

// getNodes returns an array of nodes at the given depth.
func getNodes(tree *Tree, depth uint) []*Node {
	nodes := make([]*Node, 0, 1<<depth)
	tree.root.traverseDepth(tree, depth, func(node *Node) {
		nodes = append(nodes, node)
	})
	return nodes
}

// call cb for every node exactly depth levels below it
// depth first search to return in tree ordering.
func (node *Node) traverseDepth(t *Tree, depth uint, cb func(*Node)) {
	// base case
	if depth == 0 {
		cb(node)
		return
	}
	if node.isLeaf() {
		return
	}

	// otherwise, descend one more level
	node.getLeftNode(t).traverseDepth(t, depth-1, cb)
	node.getRightNode(t).traverseDepth(t, depth-1, cb)
}

// position to key can calculate the appropriate sort order
// for the count-th node at a given depth, assuming a full
// tree above this height.
func positionToKey(depth, count uint) (key uint64) {
	for d := depth; d > 0; d-- {
		// lowest digit of count * 2^(d-1)
		key += uint64((count & 1) << (d - 1))
		count = count >> 1
	}
	return
}

// GetChunk finds the count-th subtree at depth and
// generates a Chunk for that data.
func GetChunk(tree *Tree, depth, count uint) Chunk {
	node := getNodes(tree, depth)[count]
	prefix := positionToKey(depth, count)
	return getChunk(tree, node, prefix, depth)
}

// getChunk takes a node and serializes all nodes below it
//
// As it is part of a larger tree, prefix defines the path
// up to this point, and depth the current depth
// (which defines where we add to the prefix)
//
// TODO: make this more efficient, *Chunk as arg???
func getChunk(t *Tree, node *Node, prefix uint64, depth uint) Chunk {
	if node.isLeaf() {
		return Chunk{NewOrderedNode(node, prefix)}
	}
	res := make(Chunk, 0, node.size)
	if node.leftNode != nil {
		left := getChunk(t, node.getLeftNode(t), prefix, depth+1)
		res = append(res, left...)
	}
	if node.rightNode != nil {
		offset := prefix + 1<<depth
		right := getChunk(t, node.getRightNode(t), offset, depth+1)
		res = append(res, right...)
	}
	return res
}

// Sort does an inline quicksort.
func (c Chunk) Sort() {
	sort.Slice(c, func(i, j int) bool {
		return c[i].SortOrder < c[j].SortOrder
	})
}

// MergeChunks does a merge sort of the two Chunks,
// assuming they were already in sorted order.
func MergeChunks(left, right Chunk) Chunk {
	size, i, j := len(left)+len(right), 0, 0
	slice := make([]OrderedNodeData, size)

	for k := 0; k < size; k++ {
		if i > len(left)-1 && j <= len(right)-1 {
			slice[k] = right[j]
			j++
		} else if j > len(right)-1 && i <= len(left)-1 {
			slice[k] = left[i]
			i++
		} else if left[i].SortOrder < right[j].SortOrder {
			slice[k] = left[i]
			i++
		} else {
			slice[k] = right[j]
			j++
		}
	}
	return Chunk(slice)
}

// CalculateRoot creates a temporary in-memory
// iavl tree to calculate the root hash of inserting
// all the nodes.
func (c Chunk) CalculateRoot() []byte {
	test := NewTree(nil, 2*len(c))
	c.PopulateTree(test)
	return test.Hash()
}

// PopulateTree adds all the chunks in order to the given tree.
func (c Chunk) PopulateTree(empty *Tree) {
	for _, data := range c {
		empty.Set(data.Key, data.Value)
	}
}
