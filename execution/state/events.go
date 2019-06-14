package state

import (
	"bytes"
	"fmt"
	"io"

	"github.com/tendermint/go-amino"

	"github.com/hyperledger/burrow/execution/exec"
)

var cdc = amino.NewCodec()

func (ws *writeState) AddBlock(be *exec.BlockExecution) error {
	// If there are no transactions, do not store anything. This reduces the amount of data we store and
	// prevents the iavl tree from changing, which means the AppHash does not change.
	if len(be.TxExecutions) == 0 {
		return nil
	}

	streamEventBytes := make([]byte, 0)

	for _, ev := range be.StreamEvents() {
		if ev.BeginTx != nil {
			val := &exec.TxExecutionKey{Height: be.Height, Offset: uint64(len(streamEventBytes))}
			bs, err := val.Encode()
			if err != nil {
				return err
			}
			// Set reference to TxExecution
			ws.plain.Set(keys.TxHash.Key(ev.BeginTx.TxHeader.TxHash), bs)
		}

		bs, err := cdc.MarshalBinaryLengthPrefixed(ev)
		if err != nil {
			return err
		}

		streamEventBytes = append(streamEventBytes, bs...)
	}

	tree, err := ws.forest.Writer(keys.Event.Prefix())
	if err != nil {
		return err
	}
	key := keys.Event.KeyNoPrefix(be.Height)
	tree.Set(key, streamEventBytes)

	return nil
}

func (s *ReadState) IterateStreamEvents(start, end *uint64, consumer func(*exec.StreamEvent) error) error {
	tree, err := s.Forest.Reader(keys.Event.Prefix())
	if err != nil {
		return err
	}
	var startKey, endKey []byte
	if start != nil {
		startKey = keys.Event.KeyNoPrefix(*start)
	}
	if end != nil {
		endKey = keys.Event.KeyNoPrefix(*end)
	}
	return tree.Iterate(startKey, endKey, true, func(_, value []byte) error {
		r := bytes.NewReader(value)

		for r.Len() > 0 {
			ev := new(exec.StreamEvent)
			_, err := cdc.UnmarshalBinaryLengthPrefixedReader(r, ev, 0)
			if err != nil {
				return err
			}

			err = consumer(ev)
			if err != nil {
				break
			}
		}

		return err
	})
}

func (s *ReadState) TxsAtHeight(height uint64) ([]*exec.TxExecution, error) {
	var stack exec.TxStack
	var txExecutions []*exec.TxExecution
	start := height
	end := height + 1
	err := s.IterateStreamEvents(&start, &end,
		func(ev *exec.StreamEvent) error {
			// Keep trying to consume TxExecutions at from events at this height
			txe := stack.Consume(ev)
			if txe != nil {
				txExecutions = append(txExecutions, txe)
			}
			return nil
		})
	if err != nil && err != io.EOF {
		return nil, err
	}
	return txExecutions, nil
}

func (s *ReadState) TxByHash(txHash []byte) (*exec.TxExecution, error) {
	const errHeader = "TxByHash():"
	bs := s.Plain.Get(keys.TxHash.Key(txHash))
	if len(bs) == 0 {
		return nil, nil
	}

	key, err := exec.DecodeTxExecutionKey(bs)
	if err != nil {
		return nil, err
	}

	blockTree, err := s.Forest.Reader(keys.Event.Prefix())
	if err != nil {
		return nil, err
	}

	bs = blockTree.Get(keys.Event.KeyNoPrefix(key.Height))
	if len(bs) == 0 {
		return nil, fmt.Errorf("%s could not retrieve transaction with TxHash %X despite finding reference",
			errHeader, txHash)
	}

	r := bytes.NewReader(bs[key.Offset:])
	var stack exec.TxStack

	for {
		ev := new(exec.StreamEvent)
		_, err := cdc.UnmarshalBinaryLengthPrefixedReader(r, ev, 0)
		if err != nil {
			return nil, err
		}

		txe := stack.Consume(ev)
		if txe != nil {
			return txe, nil
		}
	}
}
