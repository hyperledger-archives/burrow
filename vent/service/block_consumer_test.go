package service

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/hyperledger/burrow/vent/chain/burrow"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/solidity"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/vent/chain"
	"github.com/hyperledger/burrow/vent/sqlsol"
	"github.com/hyperledger/burrow/vent/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

const chainID = "TestChainID"

func TestBlockConsumer(t *testing.T) {
	doneCh := make(chan struct{})
	eventCh := make(chan types.EventData, 100)

	spec, err := abi.ReadSpec(solidity.Abi_EventEmitter)
	require.NoError(t, err)

	type args struct {
		Direction []byte
		Trueism   bool
		German    string
		NewDepth  *big.Int
		Bignum    int8
		Hash      string
	}
	manyTypesEventSpec := spec.EventsByName["ManyTypes"]

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
	topics, data, err := abi.PackEvent(manyTypesEventSpec, in)
	require.NoError(t, err)
	log := &exec.LogEvent{
		Address: crypto.Address{},
		Data:    data,
		Topics:  topics,
	}

	logger := logging.NewNoopLogger()
	fieldMappings := []*types.EventFieldMapping{
		{
			Field:         "direction",
			Type:          types.EventFieldTypeString,
			ColumnName:    "direction",
			BytesToString: true,
		},
	}
	t.Run("Consume matching event", func(t *testing.T) {
		spec, err := abi.ReadSpec(solidity.Abi_EventEmitter)
		require.NoError(t, err)

		longFilter := "(Log1Text = 'a' OR Log1Text = 'b' OR Log1Text = 'c' OR Log1Text = 'frogs') AND EventName = 'ManyTypes'"
		require.True(t, len(longFilter) > 100)
		tableName := "Events"
		projection, err := sqlsol.NewProjection(types.ProjectionSpec{
			{
				TableName:     tableName,
				Filter:        longFilter,
				FieldMappings: fieldMappings,
			},
		})
		require.NoError(t, err)
		blockConsumer := NewBlockConsumer(chainID, projection, sqlsol.None, spec.GetEventAbi, eventCh, doneCh, logger)
		tables, err := consumeBlock(blockConsumer, eventCh, log)
		require.NoError(t, err)
		rows := tables[tableName]
		assert.Len(t, rows, 1)
		assert.Equal(t, direction, rows[0].RowData["direction"])
	})

	t.Run("Consume matching event without ABI", func(t *testing.T) {
		spec, err := abi.ReadSpec(solidity.Abi_EventEmitter)
		require.NoError(t, err)

		// Remove the ABI
		delete(spec.EventsByID, manyTypesEventSpec.ID)

		tableName := "Events"
		// Here we are using a filter that matches - but we no longer have ABI
		projection, err := sqlsol.NewProjection(types.ProjectionSpec{
			{
				TableName:     tableName,
				Filter:        "Log1Text = 'a' OR Log1Text = 'b' OR Log1Text = 'frogs'",
				FieldMappings: fieldMappings,
			},
		})
		require.NoError(t, err)
		blockConsumer := NewBlockConsumer(chainID, projection, sqlsol.None, spec.GetEventAbi, eventCh, doneCh, logger)
		_, err = consumeBlock(blockConsumer, eventCh, log)
		require.Error(t, err)
		require.Contains(t, err.Error(), "could not find ABI")
	})

	t.Run("Consume non-matching event without ABI", func(t *testing.T) {
		spec, err := abi.ReadSpec(solidity.Abi_EventEmitter)
		require.NoError(t, err)

		// Remove the ABI
		delete(spec.EventsByID, manyTypesEventSpec.ID)

		tableName := "Events"
		// Here we are using a filter that matches - but we no longer have ABI
		projection, err := sqlsol.NewProjection(types.ProjectionSpec{
			{
				TableName:     tableName,
				Filter:        "ThisIsNotAKey = 'bar'",
				FieldMappings: fieldMappings,
			},
		})
		require.NoError(t, err)
		blockConsumer := NewBlockConsumer(chainID, projection, sqlsol.None, spec.GetEventAbi, eventCh, doneCh, logger)
		table, err := consumeBlock(blockConsumer, eventCh, log)
		require.Len(t, table, 0, "should match no event")
	})

	// This is possibly 'bad' behaviour - since you may be missing an ABI - but for now it is expected. On-chain ABIs
	// ought to solve this
	t.Run("Consume event that doesn't match without ABI tags", func(t *testing.T) {
		// This is the case where we silently fail - it would match if we had the ABI - but since we don't it doesn't
		// and we just carry on
		tableName := "Events"
		// Here we are using a filter that matches - but we no longer have ABI
		projection, err := sqlsol.NewProjection(types.ProjectionSpec{
			{
				TableName:     tableName,
				Filter:        "EventName = 'ManyTypes'",
				FieldMappings: fieldMappings,
			},
		})
		require.NoError(t, err)

		spec, err := abi.ReadSpec(solidity.Abi_EventEmitter)
		require.NoError(t, err)

		blockConsumer := NewBlockConsumer(chainID, projection, sqlsol.None, spec.GetEventAbi, eventCh, doneCh, logger)
		table, err := consumeBlock(blockConsumer, eventCh, log)
		// Check matches
		require.NoError(t, err)
		require.Len(t, table, 1)
		require.Len(t, table[tableName], 1)
		// Now Remove the ABI - should not match the event
		delete(spec.EventsByID, manyTypesEventSpec.ID)
		blockConsumer = NewBlockConsumer(chainID, projection, sqlsol.None, spec.GetEventAbi, eventCh, doneCh, logger)
		table, err = consumeBlock(blockConsumer, eventCh, log)
		require.NoError(t, err)
		require.Len(t, table, 0, "should match no events")
	})
}

const timeout = time.Second

var errTimeout = fmt.Errorf("timed out after %s waiting for consumer to emit block event", timeout)

func consumeBlock(blockConsumer func(block chain.Block) error, eventCh <-chan types.EventData,
	logEvents ...*exec.LogEvent) (map[string]types.EventDataTable, error) {

	block := &exec.BlockExecution{
		Header: &tmproto.Header{},
	}
	for _, logEvent := range logEvents {
		txe := &exec.TxExecution{
			TxHeader: &exec.TxHeader{},
		}
		err := txe.Log(logEvent)
		if err != nil {
			return nil, err
		}
		block.AppendTxs(txe)
	}
	err := blockConsumer(burrow.NewBurrowBlock(block))
	if err != nil {
		return nil, err
	}
	select {
	case <-time.After(timeout):
		return nil, errTimeout
	case ed := <-eventCh:
		return ed.Tables, nil
	}
}
