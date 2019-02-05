package state

import (
	"fmt"

	"github.com/hyperledger/burrow/execution/exec"
)

func (ws *writeState) AddBlock(be *exec.BlockExecution) error {
	tree, err := ws.forest.Writer(keys.Event.Prefix())
	if err != nil {
		return err
	}
	txHashTree, err := ws.forest.Writer(keys.TxHash.Prefix())
	if err != nil {
		return err
	}
	// Index transactions so they can be retrieved by their TxHash
	for _, ev := range be.Events() {
		key := keys.Event.KeyNoPrefix(be.Height, ev.Index)
		bs, err := ev.Encode()
		if err != nil {
			return err
		}
		// Set StreamEvent itself
		tree.Set(key, bs)
		if ev.TxExecution != nil {
			// Set reference to TxExecution
			txHashTree.Set(keys.TxHash.KeyNoPrefix(ev.TxExecution.TxHash), key)
		}
	}
	return nil
}

func (s *ReadState) IterateStreamEvents(startHeight, endHeight uint64, consumer func(*exec.StreamEvent) error) error {
	tree, err := s.Forest.Reader(keys.Event.Prefix())
	if err != nil {
		return err
	}
	return tree.Iterate(keys.Event.KeyNoPrefix(startHeight), keys.Event.KeyNoPrefix(endHeight), true,
		func(_, value []byte) error {
			txe, err := exec.DecodeBlockEvent(value)
			if err != nil {
				return fmt.Errorf("error unmarshalling BlockExecution in GetBlocks: %v", err)
			}
			return consumer(txe)
		})
}

func (s *ReadState) TxsAtHeight(height uint64) ([]*exec.TxExecution, error) {
	var txExecutions []*exec.TxExecution
	err := s.IterateStreamEvents(height, height+1, func(ev *exec.StreamEvent) error {
		if ev.TxExecution != nil {
			txExecutions = append(txExecutions, ev.TxExecution)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return txExecutions, nil
}

func (s *ReadState) StreamEvent(height, index uint64) (*exec.StreamEvent, error) {
	tree, err := s.Forest.Reader(keys.Event.Prefix())
	if err != nil {
		return nil, err
	}
	bs := tree.Get(keys.Event.KeyNoPrefix(height, index))
	if len(bs) == 0 {
		return nil, nil
	}
	return exec.DecodeBlockEvent(bs)
}

func (s *ReadState) TxByHash(txHash []byte) (*exec.TxExecution, error) {
	txHashKey, err := s.Forest.Reader(keys.TxHash.Prefix())
	if err != nil {
		return nil, err
	}
	key := txHashKey.Get(keys.TxHash.KeyNoPrefix(txHash))
	if len(key) == 0 {
		return nil, nil
	}
	tree, err := s.Forest.Reader(keys.Event.Prefix())
	if err != nil {
		return nil, err
	}
	bs := tree.Get(key)
	if len(bs) == 0 {
		return nil, fmt.Errorf("could not retieve transaction with TxHash %X despite finding reference", txHash)
	}
	ev, err := exec.DecodeBlockEvent(bs)
	if err != nil {
		return nil, err
	}
	if ev.TxExecution == nil {
		return nil, fmt.Errorf("StreamEvent %v is not a transaction despite being indexed as such", ev)
	}
	return ev.TxExecution, nil
}
