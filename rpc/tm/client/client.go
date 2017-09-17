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

	acm "github.com/hyperledger/burrow/account"
	rpc_types "github.com/hyperledger/burrow/rpc/tm/types"
	"github.com/hyperledger/burrow/txs"
	"github.com/tendermint/go-wire"
	"github.com/hyperledger/burrow/execution"
)

type RPCClient interface {
	Call(method string, params map[string]interface{}, result interface{}) (interface{}, error)
}

func Status(client RPCClient) (*rpc_types.ResultStatus, error) {
	res, err := call(client, "status")
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultStatus), nil
}

func ChainId(client RPCClient) (*rpc_types.ResultChainId, error) {
	res, err := call(client, "chain_id")
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultChainId), nil
}

func GenPrivAccount(client RPCClient) (*acm.ConcretePrivateAccount, error) {
	res, err := call(client, "unsafe/gen_priv_account")
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultGenPrivAccount).PrivAccount, nil
}

func GetAccount(client RPCClient, address acm.Address) (*acm.ConcreteAccount, error) {
	res, err := call(client, "get_account",
		"address", address)
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultGetAccount).Account, nil
}

func SignTx(client RPCClient, tx txs.Tx,
	privAccounts []*acm.ConcretePrivateAccount) (txs.Tx, error) {
	res, err := call(client, "unsafe/sign_tx",
		"tx", wrappedTx{tx},
		"privAccounts", privAccounts)
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultSignTx).Tx, nil
}

func BroadcastTx(client RPCClient,
	tx txs.Tx) (txs.Receipt, error) {
	res, err := call(client, "broadcast_tx",
		"tx", wrappedTx{tx})
	if err != nil {
		return txs.Receipt{}, err
	}
	receiptBytes := res.(*rpc_types.ResultBroadcastTx).Data
	receipt := txs.Receipt{}
	err = wire.ReadBinaryBytes(receiptBytes, &receipt)
	return receipt, err

}

func DumpStorage(client RPCClient,
	address acm.Address) (*rpc_types.ResultDumpStorage, error) {
	res, err := call(client, "dump_storage",
		"address", address)
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultDumpStorage), err
}

func GetStorage(client RPCClient, address acm.Address, key []byte) ([]byte, error) {
	res, err := call(client, "get_storage",
		"address", address,
		"key", key)
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultGetStorage).Value, nil
}

func CallCode(client RPCClient, fromAddress, code,
	data []byte) (*rpc_types.ResultCall, error) {
	res, err := call(client, "call_code",
		"fromAddress", fromAddress,
		"code", code,
		"data", data)
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultCall), err
}

func Call(client RPCClient, fromAddress, toAddress acm.Address,
	data []byte) (*rpc_types.ResultCall, error) {
	res, err := call(client, "call",
		"fromAddress", fromAddress,
		"toAddress", toAddress,
		"data", data)
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultCall), err
}

func GetName(client RPCClient, name string) (*execution.NameRegEntry, error) {
	res, err := call(client, "get_name",
		"name", name)
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultGetName).Entry, nil
}

func BlockchainInfo(client RPCClient, minHeight,
	maxHeight int) (*rpc_types.ResultBlockchainInfo, error) {
	res, err := call(client, "blockchain",
		"minHeight", minHeight,
		"maxHeight", maxHeight)
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultBlockchainInfo), err
}

func GetBlock(client RPCClient, height int) (*rpc_types.ResultGetBlock, error) {
	res, err := call(client, "get_block",
		"height", height)
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultGetBlock), err
}

func ListUnconfirmedTxs(client RPCClient) (*rpc_types.ResultListUnconfirmedTxs, error) {
	res, err := call(client, "list_unconfirmed_txs")
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultListUnconfirmedTxs), err
}

func ListValidators(client RPCClient) (*rpc_types.ResultListValidators, error) {
	res, err := call(client, "list_validators")
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultListValidators), err
}

func DumpConsensusState(client RPCClient) (*rpc_types.ResultDumpConsensusState, error) {
	res, err := call(client, "dump_consensus_state")
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultDumpConsensusState), err
}

func call(client RPCClient, method string,
	paramKeyVals ...interface{}) (res rpc_types.BurrowResult, err error) {
	pMap, err := paramsMap(paramKeyVals...)
	if err != nil {
		return
	}
	_, err = client.Call(method, pMap, &res)
	return
}

func paramsMap(orderedKeyVals ...interface{}) (map[string]interface{}, error) {
	if len(orderedKeyVals)%2 != 0 {
		return nil, fmt.Errorf("mapAndValues requires a even length list of"+
			" keys and values but got: %v (length %v)",
			orderedKeyVals, len(orderedKeyVals))
	}
	paramsMap := make(map[string]interface{})
	for i := 0; i < len(orderedKeyVals); i += 2 {
		key, ok := orderedKeyVals[i].(string)
		if !ok {
			return nil, errors.New("mapAndValues requires every even element" +
				" of orderedKeyVals to be a string key")
		}
		val := orderedKeyVals[i+1]
		paramsMap[key] = val
	}
	return paramsMap, nil
}

type wrappedTx struct {
	txs.Tx `json:"unwrap"`
}
