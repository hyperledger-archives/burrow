package abci

import (
	"fmt"

	"github.com/hyperledger/burrow/consensus/tendermint/codes"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/txs"
	"github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/common"
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
	tags := []common.KVPair{{Key: []byte(structure.TxHashKey), Value: []byte(txEnv.Tx.Hash().String())}}

	txe, err := executor.Execute(txEnv)
	if err != nil {
		ex := errors.AsException(err)
		return types.ResponseCheckTx{
			Code: codes.TxExecutionErrorCode,
			Tags: tags,
			Log:  logf("Could not execute transaction: %s, error: %v", txEnv, ex.Exception),
		}
	}

	if txe.Receipt.CreatesContract {
		tags = append(tags, common.KVPair{
			Key:   []byte("created_contract_address"),
			Value: []byte(txe.Receipt.ContractAddress.String()),
		})
	}

	bs, err := txe.Receipt.Encode()
	if err != nil {
		return types.ResponseCheckTx{
			Code: codes.EncodingErrorCode,
			Tags: tags,
			Log:  logf("Could not serialise receipt: %s", err),
		}
	}
	return types.ResponseCheckTx{
		Code: codes.TxExecutionSuccessCode,
		Tags: tags,
		Log:  logf("Execution success - TxExecution in data"),
		Data: bs,
	}
}

// Some ABCI type helpers

func WithTags(logger *logging.Logger, tags []common.KVPair) *logging.Logger {
	keyvals := make([]interface{}, len(tags)*2)
	for i, kvp := range tags {
		keyvals[i] = string(kvp.Key)
		keyvals[i+1] = string(kvp.Value)
	}
	return logger.With(keyvals...)
}

func DeliverTxFromCheckTx(ctr types.ResponseCheckTx) types.ResponseDeliverTx {
	return types.ResponseDeliverTx{
		Code:      ctr.Code,
		Log:       ctr.Log,
		Data:      ctr.Data,
		Tags:      ctr.Tags,
		GasUsed:   ctr.GasUsed,
		GasWanted: ctr.GasWanted,
		Info:      ctr.Info,
	}
}
