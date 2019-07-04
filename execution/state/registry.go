package state

import (
	"fmt"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/encoding"
	"github.com/hyperledger/burrow/execution/registry"
)

var _ registry.IterableReader = &State{}

// GetNetworkRegistry returns for each validator address, the list of their identified node at the current state
func (s *ReadState) GetNetworkRegistry() (map[crypto.Address]*registry.RegisteredNode, error) {
	netReg := make(map[crypto.Address]*registry.RegisteredNode)
	err := s.IterateNodes(func(addr crypto.Address, reg *registry.RegisteredNode) error {
		netReg[addr] = reg
		return nil
	})
	return netReg, err
}

func (s *ReadState) IterateNodes(consumer func(crypto.Address, *registry.RegisteredNode) error) error {
	tree, err := s.Forest.Reader(keys.Registry.Prefix())
	if err != nil {
		return err
	}
	return tree.Iterate(nil, nil, true, func(key []byte, value []byte) error {
		rn := new(registry.RegisteredNode)
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

func (ws *writeState) RegisterNode(val crypto.Address, regNode *registry.RegisteredNode) error {
	if regNode == nil {
		return fmt.Errorf("RegisterNode passed nil node in State")
	}

	bs, err := encoding.Encode(regNode)
	if err != nil {
		return fmt.Errorf("RegisterNode could not encode node: %v", err)
	}
	tree, err := ws.forest.Writer(keys.Registry.Prefix())
	if err != nil {
		return err
	}

	tree.Set(keys.Registry.KeyNoPrefix(val), bs)
	return nil
}
