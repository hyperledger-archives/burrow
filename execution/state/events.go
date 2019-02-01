package state

import (
	"fmt"

	"github.com/hyperledger/burrow/execution/exec"
)

// Execution events
func (ws *writeState) AddTxs(txExecutions []*exec.TxExecution) error {
	txTree, err := ws.forest.Writer(keys.Tx.Prefix())
	if err != nil {
		return err
	}
	txHashTree, err := ws.forest.Writer(keys.TxHash.Prefix())
	if err != nil {
		return err
	}
	// Index transactions so they can be retrieved by their TxHash
	for _, txe := range txExecutions {
		key := keys.Tx.KeyNoPrefix(txe.Height, txe.Index)
		bs, err := txe.Encode()
		if err != nil {
			return err
		}
		// Set TxExecution itself
		txTree.Set(key, bs)
		// Set reference to it
		txHashTree.Set(txe.TxHash, key)
	}
	return nil
}

func (s *ReadState) GetTxs(startHeight, endHeight uint64, consumer func(*exec.TxExecution) error) error {
	tree, err := s.Forest.Reader(keys.Tx.Prefix())
	if err != nil {
		return err
	}
	return tree.Iterate(keys.Tx.KeyNoPrefix(startHeight), keys.Tx.KeyNoPrefix(endHeight), true,
		func(_, value []byte) error {
			txe, err := exec.DecodeTxExecution(value)
			if err != nil {
				return fmt.Errorf("error unmarshalling BlockExecution in GetBlocks: %v", err)
			}
			return consumer(txe)
		})
}

func (s *ReadState) GetTxsAtHeight(height uint64) ([]*exec.TxExecution, error) {
	var txExecutions []*exec.TxExecution
	err := s.GetTxs(height, height+1, func(txe *exec.TxExecution) error {
		txExecutions = append(txExecutions, txe)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return txExecutions, nil
}

func (s *ReadState) GetTx(height, index uint64) (*exec.TxExecution, error) {
	tree, err := s.Forest.Reader(keys.Tx.Prefix())
	if err != nil {
		return nil, err
	}
	bs := tree.Get(keys.Tx.KeyNoPrefix(height, index))
	if len(bs) == 0 {
		return nil, nil
	}
	return exec.DecodeTxExecution(bs)
}

func (s *ReadState) GetTxByHash(txHash []byte) (*exec.TxExecution, error) {
	txHashKey, err := s.Forest.Reader(keys.TxHash.Prefix())
	if err != nil {
		return nil, err
	}
	key := txHashKey.Get(keys.TxHash.KeyNoPrefix(txHash))
	if len(key) == 0 {
		return nil, nil
	}
	txTree, err := s.Forest.Reader(keys.Tx.Prefix())
	if err != nil {
		return nil, err
	}
	bs := txTree.Get(key)
	if len(bs) == 0 {
		return nil, fmt.Errorf("could not retieve transaction with TxHash %X despite finding reference", txHash)
	}
	return exec.DecodeTxExecution(bs)
}
