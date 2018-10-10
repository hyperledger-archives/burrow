package exec

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/txs"
	abciTypes "github.com/tendermint/tendermint/abci/types"
)

func EventStringBlockExecution(height uint64) string { return fmt.Sprintf("Execution/Block/%v", height) }

func DecodeBlockExecution(bs []byte) (*BlockExecution, error) {
	be := new(BlockExecution)
	err := cdc.UnmarshalBinary(bs, be)
	if err != nil {
		return nil, err
	}
	return be, nil
}

func (be *BlockExecution) Encode() ([]byte, error) {
	return cdc.MarshalBinary(be)
}

func (*BlockExecution) EventType() EventType {
	return TypeBlockExecution
}

func (be *BlockExecution) Tx(txEnv *txs.Envelope, tx *txs.Tx) *TxExecution {
	txe := NewTxExecution(txEnv, tx)
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
			query.MustReflectTags(be.BlockHeader),
		),
		BlockExecution: be,
	}
}

type ABCIHeader struct {
	*abciTypes.Header
}

// Gogo proto support
func (h *ABCIHeader) Marshal() ([]byte, error) {
	return proto.Marshal(h.Header)
}

func (h *ABCIHeader) Unmarshal(data []byte) error {
	return proto.Unmarshal(data, h.Header)
}

func QueryForBlockExecutionFromHeight(height uint64) *query.Builder {
	return QueryForBlockExecution().AndGreaterThanOrEqual(event.HeightKey, height)
}

func QueryForBlockExecution() *query.Builder {
	return query.NewBuilder().AndEquals(event.EventTypeKey, TypeBlockExecution)
}
