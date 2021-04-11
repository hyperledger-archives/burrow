package state

import (
	"fmt"

	"github.com/hyperledger/burrow/storage"

	"github.com/hyperledger/burrow/encoding"
	"github.com/hyperledger/burrow/execution/names"
)

var _ names.IterableReader = &State{}

func (s *ImmutableState) GetName(name string) (*names.Entry, error) {
	tree, err := s.Forest.Reader(keys.Name.Prefix())
	if err != nil {
		return nil, err
	}
	entryBytes, err := tree.Get(keys.Name.KeyNoPrefix(name))
	if err != nil {
		return nil, err
	} else if entryBytes == nil {
		return nil, nil
	}

	entry := new(names.Entry)
	return entry, encoding.Decode(entryBytes, entry)
}

func (ws *writeState) UpdateName(entry *names.Entry) error {
	return ws.forest.Write(keys.Name.Prefix(), func(tree *storage.RWTree) error {
		bs, err := encoding.Encode(entry)
		if err != nil {
			return err
		}
		tree.Set(keys.Name.KeyNoPrefix(entry.Name), bs)
		return nil
	})
}

func (ws *writeState) RemoveName(name string) error {
	return ws.forest.Write(keys.Name.Prefix(), func(tree *storage.RWTree) error {
		tree.Delete(keys.Name.KeyNoPrefix(name))
		return nil
	})
}

func (s *ImmutableState) IterateNames(consumer func(*names.Entry) error) error {
	tree, err := s.Forest.Reader(keys.Name.Prefix())
	if err != nil {
		return err
	}
	return tree.Iterate(nil, nil, true, func(key []byte, value []byte) error {
		entry := new(names.Entry)
		err := encoding.Decode(value, entry)
		if err != nil {
			return fmt.Errorf("State.IterateNames() could not iterate over names: %v", err)
		}
		return consumer(entry)
	})
}
