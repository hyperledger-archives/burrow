package state

import (
	"fmt"

	"github.com/hyperledger/burrow/execution/names"
)

var _ names.IterableReader = &State{}

func (s *ReadState) GetName(name string) (*names.Entry, error) {
	tree, err := s.forest.Reader(nameKeyFormat.Prefix())
	if err != nil {
		return nil, err
	}
	entryBytes := tree.Get(nameKeyFormat.KeyNoPrefix(name))
	if entryBytes == nil {
		return nil, nil
	}

	return names.DecodeEntry(entryBytes)
}

func (ws *writeState) UpdateName(entry *names.Entry) error {
	tree, err := ws.forest.Writer(nameKeyFormat.Prefix())
	if err != nil {
		return err
	}
	bs, err := entry.Encode()
	if err != nil {
		return err
	}
	tree.Set(nameKeyFormat.KeyNoPrefix(entry.Name), bs)
	return nil
}

func (ws *writeState) RemoveName(name string) error {
	tree, err := ws.forest.Writer(nameKeyFormat.Prefix())
	if err != nil {
		return err
	}
	tree.Delete(nameKeyFormat.KeyNoPrefix(name))
	return nil
}

func (s *ReadState) IterateNames(consumer func(*names.Entry) error) error {
	tree, err := s.forest.Reader(nameKeyFormat.Prefix())
	if err != nil {
		return err
	}
	return tree.Iterate(nil, nil, true, func(key []byte, value []byte) error {
		entry, err := names.DecodeEntry(value)
		if err != nil {
			return fmt.Errorf("State.IterateNames() could not iterate over names: %v", err)
		}
		return consumer(entry)
	})
}
