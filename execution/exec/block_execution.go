package exec

import (
	"fmt"

	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/txs"
)

func EventStringBlockExecution(height uint64) string { return fmt.Sprintf("Execution/Block/%v", height) }

func DecodeBlockEvent(bs []byte) (*BlockEvent, error) {
	be := new(BlockEvent)
	err := cdc.UnmarshalBinaryBare(bs, be)
	if err != nil {
		return nil, err
	}
	return be, nil
}

// Write out TxExecutions parenthetically
func (be *BlockExecution) Events() []*BlockEvent {
	evs := make([]*BlockEvent, len(be.TxExecutions)+2)
	evs[0] = &BlockEvent{
		Index: 0,
		BeginBlock: &BeginBlock{
			Height: be.Height,
			Header: be.Header,
		},
	}
	for i, txe := range be.TxExecutions {
		evs[i+1] = &BlockEvent{
			Index:       uint64(i + 1),
			TxExecution: txe,
		}
	}
	end := len(evs) - 1
	evs[end] = &BlockEvent{
		Index: uint64(end),
		EndBlock: &EndBlock{
			Height: be.Height,
		},
	}
	return evs
}

func (be *BlockExecution) Encode() ([]byte, error) {
	return cdc.MarshalBinaryBare(be)
}

func (be *BlockExecution) EncodeHeader() ([]byte, error) {
	return cdc.MarshalBinaryBare(be.Header)
}

func (be *BlockEvent) Encode() ([]byte, error) {
	return cdc.MarshalBinaryBare(be)
}

func (*BlockExecution) EventType() EventType {
	return TypeBlockExecution
}

func (be *BlockExecution) Tx(txEnv *txs.Envelope) *TxExecution {
	txe := NewTxExecution(txEnv)
	be.Append(txe)
	return txe
}

func (be *BlockExecution) Append(tail ...*TxExecution) {
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
	*BlockEvent
}

func (ev *BlockEvent) EventType() EventType {
	switch {
	case ev.BeginBlock != nil:
		return TypeBeginBlock
	case ev.TxExecution != nil:
		return TypeTxExecution
	case ev.EndBlock != nil:
		return TypeEndBlock
	}
	return TypeUnknown
}

func (ev *BlockEvent) Tagged() *TaggedBlockEvent {
	return &TaggedBlockEvent{
		Tagged: query.MergeTags(
			ev.TxExecution.Tagged(),
			query.TagMap{
				event.EventTypeKey: ev.EventType(),
			},
			query.MustReflectTags(ev.BeginBlock, "Height"),
			query.MustReflectTags(ev.BeginBlock.GetHeader()),
			query.MustReflectTags(ev.EndBlock, "Height"),
		),
		BlockEvent: ev,
	}
}
