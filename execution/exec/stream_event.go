package exec

import (
	"io"
)
type EventStream interface {
	Recv() (*StreamEvent, error)
}

type StreamEvents []*StreamEvent

func (ses *StreamEvents) Recv() (*StreamEvent, error) {
	evs := *ses
	if len(evs) == 0 {
		return nil, io.EOF
	}
	ev := evs[0]
	*ses = evs[1:]
	return ev, nil
}

func ConsumeBlockExecution(stream EventStream) (block *BlockExecution, err error) {
	var ev *StreamEvent
	accum := new(BlockAccumulator)
	for ev, err = stream.Recv(); err == nil; ev, err = stream.Recv() {
		block = accum.Consume(ev)
		if block != nil {
			return block, nil
		}
	}
	// If we reach here then we have failed to consume a complete block
	return nil, err
}

type BlockAccumulator struct {
	block *BlockExecution
	stack TxStack
}

// Consume will add the StreamEvent passed to the block accumulator and if the block complete is complete return the
// BlockExecution, otherwise will return nil
func (ba *BlockAccumulator) Consume(ev *StreamEvent) *BlockExecution {
	switch {
	case ev.BeginBlock != nil:
		ba.block = &BlockExecution{
			Height: ev.BeginBlock.Height,
			Header: ev.BeginBlock.Header,
		}
	case ev.BeginTx != nil, ev.Envelope != nil, ev.Event != nil, ev.EndTx != nil:
		txe := ba.stack.Consume(ev)
		if txe != nil {
			ba.block.TxExecutions = append(ba.block.TxExecutions, txe)
		}
	case ev.EndBlock != nil:
		return ba.block
	}
	return nil
}

// TxStack is able to consume potentially nested txs
type TxStack []*TxExecution

func (stack TxStack) Peek() *TxExecution {
	return stack[len(stack)-1]
}

func (stack *TxStack) Push(txe *TxExecution) {
	// Put this txe in the parent position
	*stack = append(*stack, txe)
}

func (stack *TxStack) Pop() *TxExecution {
	s := *stack
	txc := s.Peek()
	*stack = s[:len(s)-1]
	return txc
}

// Consume will add the StreamEvent to the transaction stack and if that completes a single outermost transaction
// returns the TxExecution otherwise will return nil
func (stack *TxStack) Consume(ev *StreamEvent) *TxExecution {
	switch {
	case ev.BeginTx != nil:
		stack.Push(initTx(ev.BeginTx))
	case ev.Envelope != nil:
		txe := stack.Peek()
		txe.Envelope = ev.Envelope
		txe.Receipt = txe.Envelope.Tx.GenerateReceipt()
	case ev.Event != nil:
		txe := stack.Peek()
		txe.Events = append(txe.Events, ev.Event)
	case ev.EndTx != nil:
		txe := stack.Pop()
		if len(*stack) == 0 {
			// This terminates the outermost transaction
			return txe
		}
		// If there is a parent tx on the stack add this tx as child
		parent := stack.Peek()
		parent.TxExecutions = append(parent.TxExecutions, txe)
	}
	return nil
}

func initTx(beginTx *BeginTx) *TxExecution {
	return &TxExecution{
		TxHeader:  beginTx.TxHeader,
		Result:    beginTx.Result,
		Exception: beginTx.Exception,
	}
}
