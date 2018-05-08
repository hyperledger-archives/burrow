package iavl

// NodeData groups together a key, value and depth.
type NodeData struct {
	Key   []byte
	Value []byte
	Depth uint8
}

// SerializeFunc is any implementation that can serialize
// an iavl Node and its descendants.
type SerializeFunc func(*Tree, *Node) []NodeData

// RestoreFunc is an implementation that can restore an iavl tree from
// NodeData.
type RestoreFunc func(*Tree, []NodeData)

// Restore will take an (empty) tree restore it
// from the keys returned from a SerializeFunc.
func Restore(empty *Tree, kvs []NodeData) {
	for _, kv := range kvs {
		empty.Set(kv.Key, kv.Value)
	}
	empty.Hash()
}

func RestoreUsingDepth(empty *Tree, kvs []NodeData) {
	// Create an array of arrays of nodes. We're going to store each depth in
	// here, forming a kind of pyramid.
	depths := [][]*Node{}

	// Go through all the leaf nodes, grouping them in pairs and creating their
	// parents recursively.
	for _, kv := range kvs {
		var (
			// Left and right nodes.
			l     *Node = nil
			r     *Node = NewNode(kv.Key, kv.Value, 1)
			depth uint8 = kv.Depth
		)
		// Create depths as needed.
		for len(depths) < int(depth)+1 {
			depths = append(depths, []*Node{})
		}
		depths[depth] = append(depths[depth], r) // Add the leaf node to this depth.

		// If the nodes at this level are uneven after adding a node to it, it
		// means we have to wait for another node to be appended before we have
		// a pair. If we do have a pair, go up the tree until we don't.
		for d := depth; len(depths[d])%2 == 0; d-- {
			nodes := depths[d] // List of nodes at this depth.

			l = nodes[len(nodes)-1-1]
			r = nodes[len(nodes)-1]

			depths[d-1] = append(depths[d-1], &Node{
				key:       leftmost(r).Key,
				height:    maxInt8(l.height, r.height) + 1,
				size:      l.size + r.size,
				leftNode:  l,
				rightNode: r,
				version:   1,
			})
		}
	}
	empty.root = depths[0][0]
	empty.Hash()
}

// InOrderSerialize returns all key-values in the
// key order (as stored). May be nice to read, but
// when recovering, it will create a different.
func InOrderSerialize(t *Tree, root *Node) []NodeData {
	res := make([]NodeData, 0, root.size)
	root.traverseWithDepth(t, true, func(node *Node, depth uint8) bool {
		if node.height == 0 {
			kv := NodeData{Key: node.key, Value: node.value, Depth: depth}
			res = append(res, kv)
		}
		return false
	})
	return res
}

// StableSerializeBFS serializes the tree in a breadth-first manner.
func StableSerializeBFS(t *Tree, root *Node) []NodeData {
	if root == nil {
		return nil
	}

	size := root.size
	visited := map[string][]byte{}
	keys := make([][]byte, 0, size)
	numKeys := -1

	// Breadth-first search. At every depth, add keys in search order. Keep
	// going as long as we find keys at that depth. When we reach a leaf, set
	// its value in the visited map.
	// Since we have an AVL+ tree, the inner nodes contain only keys and not
	// values, while the leaves contain both. Note also that there are N-1 inner
	// nodes for N keys, so one of the leaf keys is only set once we reach the leaves
	// of the tree.
	for depth := uint(0); len(keys) > numKeys; depth++ {
		numKeys = len(keys)
		root.traverseDepth(t, depth, func(node *Node) {
			if _, ok := visited[string(node.key)]; !ok {
				keys = append(keys, node.key)
				visited[string(node.key)] = nil
			}
			if node.isLeaf() {
				visited[string(node.key)] = node.value
			}
		})
	}

	nds := make([]NodeData, size)
	for i, k := range keys {
		nds[i] = NodeData{k, visited[string(k)], 0}
	}
	return nds
}

// StableSerializeFrey exports the key value pairs of the tree
// in an order, such that when Restored from those keys, the
// new tree would have the same structure (and thus same
// shape) as the original tree.
//
// the algorithm is basically this: take the leftmost node
// of the left half and the leftmost node of the righthalf.
// Then go down a level...
// each time adding leftmost node of the right side.
// (bredth first search)
//
// Imagine 8 nodes in a balanced tree, split in half each time
// 1
// 1, 5
// 1, 5, 3, 7
// 1, 5, 3, 7, 2, 4, 6, 8
func StableSerializeFrey(t *Tree, top *Node) []NodeData {
	if top == nil {
		return nil
	}
	size := top.size

	// store all pending nodes for depth-first search
	queue := make([]*Node, 0, size)
	queue = append(queue, top)

	// to store all results - started with
	res := make([]NodeData, 0, size)
	left := leftmost(top)
	if left != nil {
		res = append(res, *left)
	}

	var n *Node
	for len(queue) > 0 {
		// pop
		n, queue = queue[0], queue[1:]

		// l := n.getLeftNode(tree)
		l := n.leftNode
		if isInner(l) {
			queue = append(queue, l)
		}

		// r := n.getRightNode(tree)
		r := n.rightNode
		if isInner(r) {
			queue = append(queue, r)
			left = leftmost(r)
			if left != nil {
				res = append(res, *left)
			}
		} else if isLeaf(r) {
			kv := NodeData{Key: r.key, Value: r.value}
			res = append(res, kv)
		}
	}

	return res
}

func isInner(n *Node) bool {
	return n != nil && !n.isLeaf()
}

func isLeaf(n *Node) bool {
	return n != nil && n.isLeaf()
}

func leftmost(node *Node) *NodeData {
	for isInner(node) {
		node = node.leftNode
	}
	if node == nil {
		return nil
	}
	return &NodeData{Key: node.key, Value: node.value}
}
