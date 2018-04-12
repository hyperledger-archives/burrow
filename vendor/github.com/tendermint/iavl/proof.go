package iavl

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ripemd160"

	"github.com/tendermint/go-wire"
	cmn "github.com/tendermint/tmlibs/common"
)

var (
	// ErrInvalidProof is returned by Verify when a proof cannot be validated.
	ErrInvalidProof = fmt.Errorf("invalid proof")

	// ErrInvalidInputs is returned when the inputs passed to the function are invalid.
	ErrInvalidInputs = fmt.Errorf("invalid inputs")

	// ErrInvalidRoot is returned when the root passed in does not match the proof's.
	ErrInvalidRoot = fmt.Errorf("invalid root")

	// ErrNilRoot is returned when the root of the tree is nil.
	ErrNilRoot = fmt.Errorf("tree root is nil")
)

type proofInnerNode struct {
	Height  int8
	Size    int64
	Version int64
	Left    []byte
	Right   []byte
}

func (n *proofInnerNode) String() string {
	return fmt.Sprintf("proofInnerNode[height=%d, ver=%d %x / %x]", n.Height, n.Version, n.Left, n.Right)
}

func (branch proofInnerNode) Hash(childHash []byte) []byte {
	hasher := ripemd160.New()
	buf := new(bytes.Buffer)
	n, err := int(0), error(nil)

	wire.WriteInt8(branch.Height, buf, &n, &err)
	wire.WriteInt64(branch.Size, buf, &n, &err)
	wire.WriteInt64(branch.Version, buf, &n, &err)

	if len(branch.Left) == 0 {
		wire.WriteByteSlice(childHash, buf, &n, &err)
		wire.WriteByteSlice(branch.Right, buf, &n, &err)
	} else {
		wire.WriteByteSlice(branch.Left, buf, &n, &err)
		wire.WriteByteSlice(childHash, buf, &n, &err)
	}
	if err != nil {
		panic(fmt.Sprintf("Failed to hash proofInnerNode: %v", err))
	}
	hasher.Write(buf.Bytes())

	return hasher.Sum(nil)
}

type proofLeafNode struct {
	KeyBytes   cmn.HexBytes `json:"key"`
	ValueBytes cmn.HexBytes `json:"value"`
	Version    int64        `json:"version"`
}

func (leaf proofLeafNode) Hash() []byte {
	hasher := ripemd160.New()
	buf := new(bytes.Buffer)
	n, err := int(0), error(nil)

	wire.WriteInt8(0, buf, &n, &err)
	wire.WriteInt64(1, buf, &n, &err)
	wire.WriteInt64(leaf.Version, buf, &n, &err)
	wire.WriteByteSlice(leaf.KeyBytes, buf, &n, &err)
	wire.WriteByteSlice(leaf.ValueBytes, buf, &n, &err)

	if err != nil {
		panic(fmt.Sprintf("Failed to hash proofLeafNode: %v", err))
	}
	hasher.Write(buf.Bytes())

	return hasher.Sum(nil)
}

func (leaf proofLeafNode) isLesserThan(key []byte) bool {
	return bytes.Compare(leaf.KeyBytes, key) == -1
}

func (leaf proofLeafNode) isGreaterThan(key []byte) bool {
	return bytes.Compare(leaf.KeyBytes, key) == 1
}

func (node *Node) pathToInnerKey(t *Tree, key []byte) (*PathToKey, *Node, error) {
	path := &PathToKey{}
	val, err := node._pathToKey(t, key, false, path)
	return path, val, err
}

func (node *Node) pathToKey(t *Tree, key []byte) (*PathToKey, *Node, error) {
	path := &PathToKey{}
	val, err := node._pathToKey(t, key, true, path)
	return path, val, err
}
func (node *Node) _pathToKey(t *Tree, key []byte, skipInner bool, path *PathToKey) (*Node, error) {
	if node.height == 0 {
		if bytes.Equal(node.key, key) {
			return node, nil
		}
		return nil, errors.New("key does not exist")
	} else if !skipInner && bytes.Equal(node.key, key) {
		return node, nil
	}

	if bytes.Compare(key, node.key) < 0 {
		if n, err := node.getLeftNode(t)._pathToKey(t, key, skipInner, path); err != nil {
			return nil, err
		} else {
			branch := proofInnerNode{
				Height:  node.height,
				Size:    node.size,
				Version: node.version,
				Left:    nil,
				Right:   node.getRightNode(t).hash,
			}
			path.InnerNodes = append(path.InnerNodes, branch)
			return n, nil
		}
	}

	if n, err := node.getRightNode(t)._pathToKey(t, key, skipInner, path); err != nil {
		return nil, err
	} else {
		branch := proofInnerNode{
			Height:  node.height,
			Size:    node.size,
			Version: node.version,
			Left:    node.getLeftNode(t).hash,
			Right:   nil,
		}
		path.InnerNodes = append(path.InnerNodes, branch)
		return n, nil
	}
}

func (t *Tree) constructKeyAbsentProof(key []byte, proof *KeyAbsentProof) error {
	// Get the index of the first key greater than the requested key, if the key doesn't exist.
	idx, val := t.Get64(key)
	if val != nil {
		return errors.Errorf("couldn't construct non-existence proof: key 0x%x exists", key)
	}

	var (
		lkey, lval []byte
		rkey, rval []byte
	)
	if idx > 0 {
		lkey, lval = t.GetByIndex64(idx - 1)
	}
	if idx <= t.Size64()-1 {
		rkey, rval = t.GetByIndex64(idx)
	}

	if lkey == nil && rkey == nil {
		return errors.New("couldn't get keys required for non-existence proof")
	}

	if lkey != nil {
		path, node, _ := t.root.pathToKey(t, lkey)
		proof.Left = &pathWithNode{
			Path: path,
			Node: proofLeafNode{lkey, lval, node.version},
		}
	}
	if rkey != nil {
		path, node, _ := t.root.pathToKey(t, rkey)
		proof.Right = &pathWithNode{
			Path: path,
			Node: proofLeafNode{rkey, rval, node.version},
		}
	}

	return nil
}

func (t *Tree) getWithProof(key []byte) (value []byte, proof *KeyExistsProof, err error) {
	if t.root == nil {
		return nil, nil, errors.WithStack(ErrNilRoot)
	}
	t.root.hashWithCount() // Ensure that all hashes are calculated.

	path, node, err := t.root.pathToKey(t, key)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not construct path to key")
	}

	proof = &KeyExistsProof{
		RootHash:  t.root.hash,
		PathToKey: path,
		Version:   node.version,
	}
	return node.value, proof, nil
}

func (t *Tree) getInnerWithProof(key []byte) (proof *InnerKeyProof, err error) {
	if t.root == nil {
		return nil, errors.WithStack(ErrNilRoot)
	}
	t.root.hashWithCount() // Ensure that all hashes are calculated.

	path, node, err := t.root.pathToInnerKey(t, key)
	if err != nil {
		return nil, errors.Wrap(err, "could not construct path to key")
	}

	proof = &InnerKeyProof{
		&KeyExistsProof{
			RootHash:  t.root.hash,
			PathToKey: path,
			Version:   node.version,
		},
	}
	return proof, nil
}

func (t *Tree) keyAbsentProof(key []byte) (*KeyAbsentProof, error) {
	if t.root == nil {
		return nil, errors.WithStack(ErrNilRoot)
	}
	t.root.hashWithCount() // Ensure that all hashes are calculated.

	proof := &KeyAbsentProof{
		RootHash: t.root.hash,
	}
	if err := t.constructKeyAbsentProof(key, proof); err != nil {
		return nil, errors.Wrap(err, "could not construct proof of non-existence")
	}
	return proof, nil
}
