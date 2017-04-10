// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"errors"
	"fmt"

	acm "github.com/monax/burrow/account"
	core_types "github.com/monax/burrow/core/types"
	rpc_types "github.com/monax/burrow/rpc/tendermint/core/types"
	"github.com/monax/burrow/txs"
	rpcclient "github.com/tendermint/go-rpc/client"
	"github.com/tendermint/go-wire"
)

func Status(client rpcclient.Client) (*rpc_types.ResultStatus, error) {
	res, err := performCall(client, "status")
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultStatus), nil
}

func ChainId(client rpcclient.Client) (*rpc_types.ResultChainId, error) {
	res, err := performCall(client, "chain_id")
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultChainId), nil
}

func GenPrivAccount(client rpcclient.Client) (*acm.PrivAccount, error) {
	res, err := performCall(client, "unsafe/gen_priv_account")
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultGenPrivAccount).PrivAccount, nil
}

func GetAccount(client rpcclient.Client, address []byte) (*acm.Account, error) {
	res, err := performCall(client, "get_account",
		"address", address)
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultGetAccount).Account, nil
}

func SignTx(client rpcclient.Client, tx txs.Tx,
	privAccounts []*acm.PrivAccount) (txs.Tx, error) {
	res, err := performCall(client, "unsafe/sign_tx",
		"tx", wrappedTx{tx},
		"privAccounts", privAccounts)
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultSignTx).Tx, nil
}

func BroadcastTx(client rpcclient.Client,
	tx txs.Tx) (txs.Receipt, error) {
	res, err := performCall(client, "broadcast_tx",
		"tx", wrappedTx{tx})
	if err != nil {
		return txs.Receipt{}, err
	}
	receiptBytes := res.(*rpc_types.ResultBroadcastTx).Data
	receipt := txs.Receipt{}
	err = wire.ReadBinaryBytes(receiptBytes, &receipt)
	return receipt, err

}

func DumpStorage(client rpcclient.Client,
	address []byte) (*rpc_types.ResultDumpStorage, error) {
	res, err := performCall(client, "dump_storage",
		"address", address)
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultDumpStorage), err
}

func GetStorage(client rpcclient.Client, address, key []byte) ([]byte, error) {
	res, err := performCall(client, "get_storage",
		"address", address,
		"key", key)
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultGetStorage).Value, nil
}

func CallCode(client rpcclient.Client, fromAddress, code,
	data []byte) (*rpc_types.ResultCall, error) {
	res, err := performCall(client, "call_code",
		"fromAddress", fromAddress,
		"code", code,
		"data", data)
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultCall), err
}

func Call(client rpcclient.Client, fromAddress, toAddress,
	data []byte) (*rpc_types.ResultCall, error) {
	res, err := performCall(client, "call",
		"fromAddress", fromAddress,
		"toAddress", toAddress,
		"data", data)
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultCall), err
}

func GetName(client rpcclient.Client, name string) (*core_types.NameRegEntry, error) {
	res, err := performCall(client, "get_name",
		"name", name)
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultGetName).Entry, nil
}

func BlockchainInfo(client rpcclient.Client, minHeight,
	maxHeight int) (*rpc_types.ResultBlockchainInfo, error) {
	res, err := performCall(client, "blockchain",
		"minHeight", minHeight,
		"maxHeight", maxHeight)
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultBlockchainInfo), err
}

func GetBlock(client rpcclient.Client, height int) (*rpc_types.ResultGetBlock, error) {
	res, err := performCall(client, "get_block",
		"height", height)
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultGetBlock), err
}

func ListUnconfirmedTxs(client rpcclient.Client) (*rpc_types.ResultListUnconfirmedTxs, error) {
	res, err := performCall(client, "list_unconfirmed_txs")
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultListUnconfirmedTxs), err
}

func ListValidators(client rpcclient.Client) (*rpc_types.ResultListValidators, error) {
	res, err := performCall(client, "list_validators")
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultListValidators), err
}

func DumpConsensusState(client rpcclient.Client) (*rpc_types.ResultDumpConsensusState, error) {
	res, err := performCall(client, "dump_consensus_state")
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultDumpConsensusState), err
}

func performCall(client rpcclient.Client, method string,
	paramKeyVals ...interface{}) (res rpc_types.ErisDBResult, err error) {
	paramsMap, paramsSlice, err := mapAndValues(paramKeyVals...)
	if err != nil {
		return
	}
	switch cli := client.(type) {
	case *rpcclient.ClientJSONRPC:
		_, err = cli.Call(method, paramsSlice, &res)
	case *rpcclient.ClientURI:
		_, err = cli.Call(method, paramsMap, &res)
	default:
		panic(fmt.Errorf("peformCall called against an unknown rpcclient.Client %v",
			cli))
	}
	return

}

func mapAndValues(orderedKeyVals ...interface{}) (map[string]interface{},
	[]interface{}, error) {
	if len(orderedKeyVals)%2 != 0 {
		return nil, nil, fmt.Errorf("mapAndValues requires a even length list of"+
			" keys and values but got: %v (length %v)",
			orderedKeyVals, len(orderedKeyVals))
	}
	paramsMap := make(map[string]interface{})
	paramsSlice := make([]interface{}, len(orderedKeyVals)/2)
	for i := 0; i < len(orderedKeyVals); i += 2 {
		key, ok := orderedKeyVals[i].(string)
		if !ok {
			return nil, nil, errors.New("mapAndValues requires every even element" +
				" of orderedKeyVals to be a string key")
		}
		val := orderedKeyVals[i+1]
		paramsMap[key] = val
		paramsSlice[i/2] = val
	}
	return paramsMap, paramsSlice, nil
}

type wrappedTx struct {
	txs.Tx `json:"unwrap"`
}
