package service

import (
	"math/big"
	"testing"
	"time"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/solidity"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/vent/sqlsol"
	"github.com/hyperledger/burrow/vent/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmTypes "github.com/tendermint/tendermint/abci/types"
)

func TestBlockConsumer(t *testing.T) {
	doneCh := make(chan struct{})
	eventCh := make(chan types.EventData, 100)
	longFilter := "(Log1Text = 'a' OR Log1Text = 'b' OR Log1Text = 'frogs') AND EventName = 'ManyTypes'"
	tableName := "Events"
	projection, err := sqlsol.NewProjection(types.ProjectionSpec{
		{
			TableName: tableName,
			Filter:    longFilter,
			FieldMappings: []*types.EventFieldMapping{
				{
					Field:         "direction",
					Type:          types.EventFieldTypeString,
					ColumnName:    "direction",
					BytesToString: true,
				},
			},
		},
	})
	require.NoError(t, err)

	spec, err := abi.ReadSpec(solidity.Abi_EventEmitter)
	require.NoError(t, err)

	blockConsumer := NewBlockConsumer(projection, sqlsol.None, spec.GetEventAbi, eventCh, doneCh, logging.NewNoopLogger())

	type args struct {
		Direction []byte
		Trueism   bool
		German    string
		NewDepth  *big.Int
		Bignum    int8
		Hash      string
	}
	eventSpec := spec.EventsByName["ManyTypes"]

	bignum := big.NewInt(1000)
	in := args{
		Direction: make([]byte, 32),
		Trueism:   false,
		German:    "foo",
		NewDepth:  bignum,
		Bignum:    100,
		Hash:      "ba",
	}
	direction := "frogs"
	copy(in.Direction, direction)
	topics, data, err := abi.PackEvent(&eventSpec, in)
	require.NoError(t, err)

	txe := &exec.TxExecution{
		TxHeader: &exec.TxHeader{},
	}
	err = txe.Log(&exec.LogEvent{
		Address: crypto.Address{},
		Data:    data,
		Topics:  topics,
	})
	require.NoError(t, err)

	block := &exec.BlockExecution{
		Header: &tmTypes.Header{},
	}
	block.AppendTxs(txe)
	err = blockConsumer(block)
	require.NoError(t, err)
	select {
	case <-time.After(time.Second * 5):
		t.Fatalf("timed out waiting for consumer to emit block event")
	case ed := <-eventCh:
		rows := ed.Tables[tableName]
		assert.Len(t, rows, 1)
		assert.Equal(t, direction, rows[0].RowData["direction"])
	}
}
