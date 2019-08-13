package exec

import (
	"fmt"
	"io"

	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/event/query"
)

type EventStream interface {
	Recv() (*StreamEvent, error)
}

func (ses *StreamEvents) Recv() (*StreamEvent, error) {
	if len(ses.StreamEvents) == 0 {
		return nil, io.EOF
	}
	ev := ses.StreamEvents[0]
	ses.StreamEvents = ses.StreamEvents[1:]
	return ev, nil
}

func (ev *StreamEvent) EventType() EventType {
	switch {
	case ev.BeginBlock != nil:
		return TypeBeginBlock
	case ev.BeginTx != nil:
		return TypeBeginTx
	case ev.Envelope != nil:
		return TypeEnvelope
	case ev.Event != nil:
		return ev.Event.EventType()
	case ev.EndTx != nil:
		return TypeEndTx
	case ev.EndBlock != nil:
		return TypeEndBlock
	}
	return TypeUnknown
}

func (ev *StreamEvent) Get(key string) (interface{}, bool) {
	switch key {
	case event.EventTypeKey:
		return ev.EventType(), true
	}
	// Flatten this sum type
	return query.TagsFor(
		ev.GetBeginBlock().GetHeader(),
		ev.BeginBlock,
		ev.GetBeginTx().GetTxHeader(),
		ev.BeginTx,
		ev.Envelope,
		ev.Event,
		ev.EndTx,
		ev.EndBlock).Get(key)
}

func ConsumeBlockExecution(stream EventStream) (block *BlockExecution, err error) {
	var ev *StreamEvent
	accum := new(BlockAccumulator)
	for ev, err = stream.Recv(); err == nil; ev, err = stream.Recv() {
		block, err = accum.Consume(ev)
		if err != nil {
			return nil, err
		}
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
func (ba *BlockAccumulator) Consume(ev *StreamEvent) (*BlockExecution, error) {
	switch {
	case ev.BeginBlock != nil:
		ba.block = &BlockExecution{
			Height: ev.BeginBlock.Height,
			Header: ev.BeginBlock.Header,
		}
	case ev.BeginTx != nil, ev.Envelope != nil, ev.Event != nil, ev.EndTx != nil:
		txe, err := ba.stack.Consume(ev)
		if err != nil {
			return nil, err
		}
		if txe != nil {
			ba.block.TxExecutions = append(ba.block.TxExecutions, txe)
		}
	case ev.EndBlock != nil:
		return ba.block, nil
	}
	return nil, nil
}

// TxStack is able to consume potentially nested txs
type TxStack []*TxExecution

func (stack *TxStack) Push(txe *TxExecution) {
	// Put this txe in the parent position
	*stack = append(*stack, txe)
}

func (stack TxStack) Peek() (*TxExecution, error) {
	if len(stack) < 1 {
		return nil, fmt.Errorf("tried to peek from an empty TxStack - might be missing essential StreamEvents")
	}
	return stack[len(stack)-1], nil
}

func (stack *TxStack) Pop() (*TxExecution, error) {
	s := *stack
	txc, err := s.Peek()
	if err != nil {
		return nil, err
	}
	*stack = s[:len(s)-1]
	return txc, nil
}

// Consume will add the StreamEvent to the transaction stack and if that completes a single outermost transaction
// returns the TxExecution otherwise will return nil
func (stack *TxStack) Consume(ev *StreamEvent) (*TxExecution, error) {
	switch {
	case ev.BeginTx != nil:
		stack.Push(initTx(ev.BeginTx))
	case ev.Envelope != nil:
		txe, err := stack.Peek()
		if err != nil {
			return nil, err
		}
		txe.Envelope = ev.Envelope
		txe.Receipt = txe.Envelope.Tx.GenerateReceipt()
	case ev.Event != nil:
		txe, err := stack.Peek()
		if err != nil {
			return nil, err
		}
		txe.Events = append(txe.Events, ev.Event)
	case ev.EndTx != nil:
		txe, err := stack.Pop()
		if err != nil {
			return nil, err
		}
		if len(*stack) == 0 {
			// This terminates the outermost transaction
			return txe, nil
		}
		// If there is a parent tx on the stack add this tx as child
		parent, err := stack.Peek()
		if err != nil {
			return nil, err
		}
		parent.TxExecutions = append(parent.TxExecutions, txe)
	}
	return nil, nil
}

func initTx(beginTx *BeginTx) *TxExecution {
	return &TxExecution{
		TxHeader:  beginTx.TxHeader,
		Result:    beginTx.Result,
		Exception: beginTx.Exception,
	}
}
