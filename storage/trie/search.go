package trie

type HeightNode struct {
	*Node
	Height      int
	ParentIndex int
}

func (tn *Node) BreadthFirstSearch(callback func(*HeightNode) error) error {
	nodes := make([]*HeightNode, len(tn.Children()))
	for i, child := range tn.Children() {
		nodes[i] = &HeightNode{Node: child}
	}
	for len(nodes) > 0 {
		// pop
		n := nodes[0]
		nodes = nodes[1:]
		// send
		err := callback(n)
		if err != nil {
			return err
		}
		// enqueue
		height := n.Height + 1
		for _, child := range n.Children() {
			nodes = append(nodes, &HeightNode{Height: height, ParentIndex: n.index, Node: child})
		}
	}
	return nil
}
