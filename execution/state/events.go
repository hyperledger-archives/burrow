package state

import (
	"fmt"

	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/storage"
	"github.com/hyperledger/burrow/txs"
)

var blockRefKeyFormat = storage.NewMustKeyFormat("b", uint64Length)
var txKeyFormat = storage.NewMustKeyFormat("t", txs.HashLength)
var txRefKeyFormat = storage.NewMustKeyFormat("", uint64Length, uint64Length)

func (ws *writeState) AddEvents(evs []*exec.Event) error {
	// TODO: unwrap blocks
	return nil
}

// Execution events
func (ws *writeState) AddBlock(be *exec.BlockExecution) error {
	// Index transactions so they can be retrieved by their TxHash
	for i, txe := range be.TxExecutions {
		err := ws.addTx(txe.TxHash, be.Height, uint64(i))
		if err != nil {
			return err
		}
	}
	bs, err := be.Encode()
	if err != nil {
		return err
	}
	tree, err := ws.forest.Writer(blockRefKeyFormat.Prefix())
	if err != nil {
		return err
	}
	tree.Set(blockRefKeyFormat.KeyNoPrefix(be.Height), bs)
	return nil
}

func (ws *writeState) addTx(txHash []byte, height, index uint64) error {
	tree, err := ws.forest.Writer(txKeyFormat.Prefix())
	if err != nil {
		return err
	}
	tree.Set(txKeyFormat.KeyNoPrefix(txHash), txRefKeyFormat.Key(height, index))
	return nil
}

func (s *State) GetTx(txHash []byte) (*exec.TxExecution, error) {
	tree, err := s.Forest.Reader(txKeyFormat.Prefix())
	if err != nil {
		return nil, err
	}
	bs := tree.Get(txKeyFormat.KeyNoPrefix(txHash))
	if len(bs) == 0 {
		return nil, nil
	}
	height, index := new(uint64), new(uint64)
	txRefKeyFormat.Scan(bs, height, index)
	be, err := s.GetBlock(*height)
	if err != nil {
		return nil, fmt.Errorf("error getting block %v containing tx %X", height, txHash)
	}
	if *index < uint64(len(be.TxExecutions)) {
		return be.TxExecutions[*index], nil
	}
	return nil, fmt.Errorf("retrieved index %v in block %v for tx %X but block only contains %v TxExecutions",
		index, height, txHash, len(be.TxExecutions))
}

func (s *State) GetBlock(height uint64) (*exec.BlockExecution, error) {
	tree, err := s.Forest.Reader(blockRefKeyFormat.Prefix())
	if err != nil {
		return nil, err
	}
	bs := tree.Get(blockRefKeyFormat.KeyNoPrefix(height))
	if len(bs) == 0 {
		return nil, nil
	}
	return exec.DecodeBlockExecution(bs)
}

func (s *State) GetBlocks(startHeight, endHeight uint64, consumer func(*exec.BlockExecution) error) error {
	tree, err := s.Forest.Reader(blockRefKeyFormat.Prefix())
	if err != nil {
		return err
	}
	kf := blockRefKeyFormat
	return tree.Iterate(kf.KeyNoPrefix(startHeight), kf.KeyNoPrefix(endHeight), true,
		func(key []byte, value []byte) error {
			block, err := exec.DecodeBlockExecution(value)
			if err != nil {
				return fmt.Errorf("error unmarshalling BlockExecution in GetBlocks: %v", err)
			}
			return consumer(block)
		})
}
