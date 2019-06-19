package exec

import (
	"fmt"

	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/txs"
)

func EventStringBlockExecution(height uint64) string { return fmt.Sprintf("Execution/Block/%v", height) }

// Write out TxExecutions parenthetically
func (be *BlockExecution) StreamEvents() []*StreamEvent {
	var ses []*StreamEvent
	ses = append(ses, &StreamEvent{
		BeginBlock: &BeginBlock{
			Height: be.Height,
			Header: be.Header,
		},
	})
	for _, txe := range be.TxExecutions {
		ses = append(ses, txe.StreamEvents()...)
	}
	return append(ses, &StreamEvent{
		EndBlock: &EndBlock{
			Height: be.Height,
		},
	})
}

func (*BlockExecution) EventType() EventType {
	return TypeBlockExecution
}

func (be *BlockExecution) Tx(txEnv *txs.Envelope) *TxExecution {
	txe := NewTxExecution(txEnv)
	be.AppendTxs(txe)
	return txe
}

func (be *BlockExecution) AppendTxs(tail ...*TxExecution) {
	for i, txe := range tail {
		txe.Index = uint64(len(be.TxExecutions) + i)
		txe.Height = be.Height
	}
	be.TxExecutions = append(be.TxExecutions, tail...)
}

// Tags
type TaggedBlockExecution struct {
	query.Tagged
	*BlockExecution
}

func (be *BlockExecution) Tagged() *TaggedBlockExecution {
	return &TaggedBlockExecution{
		Tagged: query.MergeTags(
			query.TagMap{
				event.EventIDKey:   EventStringBlockExecution(be.Height),
				event.EventTypeKey: be.EventType(),
			},
			query.MustReflectTags(be),
			query.MustReflectTags(be.Header),
		),
		BlockExecution: be,
	}
}

func QueryForBlockExecutionFromHeight(height uint64) *query.Builder {
	return QueryForBlockExecution().AndGreaterThanOrEqual(event.HeightKey, height)
}

func QueryForBlockExecution() *query.Builder {
	return query.NewBuilder().AndEquals(event.EventTypeKey, TypeBlockExecution)
}

type TaggedBlockEvent struct {
	query.Tagged
	*StreamEvent
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

func (ev *StreamEvent) Tagged() *TaggedBlockEvent {
	return &TaggedBlockEvent{
		Tagged: query.MergeTags(
			query.TagMap{
				event.EventTypeKey: ev.EventType(),
			},
			query.MustReflectTags(ev.BeginBlock, "Height"),
			query.MustReflectTags(ev.BeginBlock.GetHeader()),
			query.MustReflectTags(ev.BeginTx),
			query.MustReflectTags(ev.BeginTx.GetTxHeader()),
			ev.Envelope.Tagged(),
			ev.Event.Tagged(),
			query.MustReflectTags(ev.EndTx),
			query.MustReflectTags(ev.EndBlock, "Height"),
		),
		StreamEvent: ev,
	}
}
