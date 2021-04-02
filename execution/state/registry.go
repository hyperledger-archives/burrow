package state

import (
	"fmt"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/encoding"
	"github.com/hyperledger/burrow/execution/registry"
	"github.com/hyperledger/burrow/storage"
)

var _ registry.IterableReader = &State{}

func getNode(forest storage.ForestReader, id crypto.Address) (*registry.NodeIdentity, error) {
	tree, err := forest.Reader(keys.Registry.Prefix())
	if err != nil {
		return nil, err
	}
	nodeBytes, err := tree.Get(keys.Registry.KeyNoPrefix(id))
	if err != nil {
		return nil, err
	} else if nodeBytes == nil {
		return nil, nil
	}

	regNode := new(registry.NodeIdentity)
	return regNode, encoding.Decode(nodeBytes, regNode)

}

func (s *ImmutableState) GetNodeByID(id crypto.Address) (*registry.NodeIdentity, error) {
	return getNode(s.Forest, id)
}

func (s *State) GetNodeIDsByAddress(net string) ([]crypto.Address, error) {
	return s.writeState.nodeStats.GetAddresses(net), nil
}

func (ws *writeState) UpdateNode(id crypto.Address, node *registry.NodeIdentity) error {
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

	prev, err := getNode(ws.forest, id)
	if err != nil {
		return err
	}

	ws.nodeStats.Remove(prev)
	ws.nodeStats.Insert(node.GetNetworkAddress(), id)
	tree.Set(keys.Registry.KeyNoPrefix(id), bs)
	return nil
}

func (ws *writeState) RemoveNode(id crypto.Address) error {
	tree, err := ws.forest.Writer(keys.Registry.Prefix())
	if err != nil {
		return err
	}

	prev, err := getNode(ws.forest, id)
	if err != nil {
		return err
	}

	ws.nodeStats.Remove(prev)
	tree.Delete(keys.Registry.KeyNoPrefix(id))
	return nil
}

func (s *ImmutableState) IterateNodes(consumer func(crypto.Address, *registry.NodeIdentity) error) error {
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

func (s *State) GetNumPeers() int {
	return len(s.writeState.nodeStats.Addresses)
}
