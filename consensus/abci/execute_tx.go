package abci

import (
	"fmt"
	"strings"

	"github.com/hyperledger/burrow/consensus/tendermint/codes"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/txs"
	"github.com/tendermint/tendermint/abci/types"
)

// Attempt to execute a transaction using ABCI conventions and codes
func ExecuteTx(logHeader string, executor execution.Executor, txDecoder txs.Decoder, txBytes []byte) types.ResponseCheckTx {
	logf := func(format string, args ...interface{}) string {
		return fmt.Sprintf("%s: "+format, append([]interface{}{logHeader}, args...)...)
	}

	txEnv, err := txDecoder.DecodeTx(txBytes)
	if err != nil {
		return types.ResponseCheckTx{
			Code: codes.EncodingErrorCode,
			Log:  logf("Decoding error: %s", err),
		}
	}
	txe, err := executor.Execute(txEnv)
	if err != nil {
		ex := errors.AsException(err)
		return types.ResponseCheckTx{
			Code: codes.TxExecutionErrorCode,
			Log:  logf("Could not execute transaction: %s, error: %v", txEnv, ex.Exception),
		}
	}

	tags := []types.EventAttribute{{Key: []byte(structure.TxHashKey), Value: []byte(txEnv.Tx.Hash().String())}}
	if txe.Receipt.CreatesContract {
		tags = append(tags, types.EventAttribute{
			Key:   []byte("created_contract_address"),
			Value: []byte(txe.Receipt.ContractAddress.String()),
		})
	}

	events := []types.Event{{Type: "ExecuteTx", Attributes: tags}}
	bs, err := txe.Receipt.Encode()
	if err != nil {
		return types.ResponseCheckTx{
			Code:   codes.EncodingErrorCode,
			Events: events,
			Log:    logf("Could not serialise receipt: %s", err),
		}
	}
	return types.ResponseCheckTx{
		Code:   codes.TxExecutionSuccessCode,
		Events: events,
		Log:    logf("Execution success - TxExecution in data"),
		Data:   bs,
	}
}

// Some ABCI type helpers

func WithEvents(logger *logging.Logger, events []types.Event) *logging.Logger {
	for _, e := range events {
		values := make([]string, 0, len(e.Attributes))
		for _, kvp := range e.Attributes {
			values = append(values, fmt.Sprintf("%s:%s", string(kvp.Key), string(kvp.Value)))
		}
		logger = logger.With(e.Type, strings.Join(values, ","))
	}
	return logger
}

func DeliverTxFromCheckTx(ctr types.ResponseCheckTx) types.ResponseDeliverTx {
	return types.ResponseDeliverTx{
		Code:      ctr.Code,
		Log:       ctr.Log,
		Data:      ctr.Data,
		Events:    ctr.Events,
		GasUsed:   ctr.GasUsed,
		GasWanted: ctr.GasWanted,
		Info:      ctr.Info,
	}
}
