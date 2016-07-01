package core

import (
	"fmt"

	ctypes "github.com/eris-ltd/eris-db/rpc/tendermint/core/types"
	txs "github.com/eris-ltd/eris-db/txs"
	// "github.com/tendermint/tendermint/types"
	tmsp "github.com/tendermint/tmsp/types"
)

//-----------------------------------------------------------------------------

// // NOTE: tx must be signed
// func BroadcastTxAsync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
// 	err := mempoolReactor.BroadcastTx(tx, nil)
// 	if err != nil {
// 		return nil, fmt.Errorf("Error broadcasting transaction: %v", err)
// 	}
// 	return &ctypes.ResultBroadcastTx{}, nil
// }

// Note: tx must be signed
func BroadcastTxSync(tx txs.Tx) (*ctypes.ResultBroadcastTx, error) {
	resCh := make(chan *tmsp.Response, 1)
	err := mempoolReactor.BroadcastTx(txs.EncodeTx(tx), func(res *tmsp.Response) {
		resCh <- res
	})
	if err != nil {
		return nil, fmt.Errorf("Error broadcasting transaction: %v", err)
	}
	res := <-resCh
	switch r:= res.Value.(type) {
	case *tmsp.Response_AppendTx:
		return &ctypes.ResultBroadcastTx{
			Code: r.AppendTx.Code,
			Data: r.AppendTx.Data,
			Log:  r.AppendTx.Log,
		}, nil
	case *tmsp.Response_CheckTx:
		return &ctypes.ResultBroadcastTx{
			Code: r.CheckTx.Code,
			Data: r.CheckTx.Data,
			Log:  r.CheckTx.Log,
		}, nil
	case *tmsp.Response_Commit:
		return &ctypes.ResultBroadcastTx{
			Code: r.Commit.Code,
			Data: r.Commit.Data,
			Log:  r.Commit.Log,
		}, nil
	case *tmsp.Response_Query:
		return &ctypes.ResultBroadcastTx{
			Code: r.Query.Code,
			Data: r.Query.Data,
			Log:  r.Query.Log,
		}, nil
	default:
		return &ctypes.ResultBroadcastTx{
			Code: tmsp.CodeType_OK,
			Data: []byte{},
			Log:  "",
		}, nil

	}
}
