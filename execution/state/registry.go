package state

import (
	"fmt"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/encoding"
	"github.com/hyperledger/burrow/execution/registry"
)

var _ registry.IterableReader = &State{}

func (s *ReadState) GetNode(addr crypto.Address) (*registry.NodeIdentity, error) {
	tree, err := s.Forest.Reader(keys.Registry.Prefix())
	if err != nil {
		return nil, err
	}
	nodeBytes := tree.Get(keys.Registry.KeyNoPrefix(addr))
	if nodeBytes == nil {
		return nil, nil
	}

	regNode := new(registry.NodeIdentity)
	return regNode, encoding.Decode(nodeBytes, regNode)
}

func (ws *writeState) UpdateNode(addr crypto.Address, node *registry.NodeIdentity) error {
	if node == nil {
		return fmt.Errorf("RegisterNode passed nil node in State")
	}

	bs, err := encoding.Encode(node)
	if err != nil {
		return fmt.Errorf("RegisterNode could not encode node: %v", err)
	}
	tree, err := ws.forest.Writer(keys.Registry.Prefix())
	if err != nil {
		return err
	}

	ws.nodeList[addr] = node
	tree.Set(keys.Registry.KeyNoPrefix(addr), bs)
	return nil
}

func (ws *writeState) RemoveNode(addr crypto.Address) error {
	tree, err := ws.forest.Writer(keys.Registry.Prefix())
	if err != nil {
		return err
	}

	delete(ws.nodeList, addr)
	tree.Delete(keys.Registry.KeyNoPrefix(addr))
	return nil
}

func (s *ReadState) IterateNodes(consumer func(crypto.Address, *registry.NodeIdentity) error) error {
	tree, err := s.Forest.Reader(keys.Registry.Prefix())
	if err != nil {
		return err
	}
	return tree.Iterate(nil, nil, true, func(key []byte, value []byte) error {
		rn := new(registry.NodeIdentity)
		err := encoding.Decode(value, rn)
		if err != nil {
			return fmt.Errorf("State.IterateNodes() could not iterate over nodes: %v", err)
		}
		address, err := crypto.AddressFromBytes(key)
		if err != nil {
			return fmt.Errorf("could not decode key: %v", err)
		}
		return consumer(address, rn)
	})
}

func (s *State) GetNodes() registry.NodeList {
	return s.writeState.nodeList
}
