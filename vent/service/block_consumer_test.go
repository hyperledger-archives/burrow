package service

import (
	"fmt"
	"testing"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/solidity"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/vent/sqlsol"
	"github.com/hyperledger/burrow/vent/types"
	"github.com/stretchr/testify/require"
	tmTypes "github.com/tendermint/tendermint/abci/types"
)

func TestBlockConsumer(t *testing.T) {
	doneCh := make(chan struct{})
	eventCh := make(chan types.EventData, 100)
	longFilter := "Log1Text = 'a' OR Log1Text = 'b'"
	projection, err := sqlsol.NewProjection(types.ProjectionSpec{
		{
			TableName: "Events",
			Filter:    longFilter,
			FieldMappings: []*types.EventFieldMapping{
				{
					Field:      "Direction",
					Type:       types.EventFieldTypeString,
					ColumnName: "direction",
				},
			},
		},
	})
	require.NoError(t, err)

	spec, err := abi.ReadSpec(solidity.Abi_EventEmitter)
	require.NoError(t, err)

	blockConsumer := NewBlockConsumer(projection, sqlsol.None, spec.GetEventAbi, eventCh, doneCh, logging.NewNoopLogger())

	type args struct {
		Direction string
		Trueism   bool
		German    string
		NewDepth  int64
		Bignum    int64
		Hash      string
	}
	eventSpec := spec.EventsByName["ManyTypes"]
	in := args{
		Direction: "b",
		Trueism:   false,
		German:    "",
		NewDepth:  100,
		Bignum:    10000,
		Hash:      "",
	}
	topics, data, err := abi.PackEvent(&eventSpec, in)
	require.NoError(t, err)

	out := new(args)
	err = abi.UnpackEvent(&eventSpec, topics, data, out)
	require.NoError(t, err)

	fmt.Println(topics)
	fmt.Println(data)

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
}
