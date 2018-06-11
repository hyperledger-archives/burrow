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
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/rpc/tm"
	"github.com/hyperledger/burrow/txs"
)

type RPCClient interface {
	Call(method string, params map[string]interface{}, result interface{}) (interface{}, error)
}

func BroadcastTx(client RPCClient, txEnv *txs.Envelope) (*txs.Receipt, error) {
	res := new(txs.Receipt)
	_, err := client.Call(tm.BroadcastTx, pmap("tx", txEnv), res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func Status(client RPCClient) (*rpc.ResultStatus, error) {
	res := new(rpc.ResultStatus)
	_, err := client.Call(tm.Status, pmap(), res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func ChainId(client RPCClient) (*rpc.ResultChainId, error) {
	res := new(rpc.ResultChainId)
	_, err := client.Call(tm.ChainID, pmap(), &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func GenPrivAccount(client RPCClient) (*rpc.ResultGeneratePrivateAccount, error) {
	res := new(rpc.ResultGeneratePrivateAccount)
	_, err := client.Call(tm.GeneratePrivateAccount, pmap(), res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func GetAccount(client RPCClient, address crypto.Address) (acm.Account, error) {
	res := new(rpc.ResultGetAccount)
	_, err := client.Call(tm.GetAccount, pmap("address", address), res)
	if err != nil {
		return nil, err
	}
	concreteAccount := res.Account
	if concreteAccount == nil {
		return nil, nil
	}
	return concreteAccount.Account(), nil
}

func SignTx(client RPCClient, tx txs.Tx, privAccounts []*acm.ConcretePrivateAccount) (*txs.Envelope, error) {
	res := new(rpc.ResultSignTx)
	_, err := client.Call(tm.SignTx, pmap("tx", tx, "privAccounts", privAccounts), res)
	if err != nil {
		return nil, err
	}
	return res.Tx, nil
}

func DumpStorage(client RPCClient, address crypto.Address) (*rpc.ResultDumpStorage, error) {
	res := new(rpc.ResultDumpStorage)
	_, err := client.Call(tm.DumpStorage, pmap("address", address), res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func GetStorage(client RPCClient, address crypto.Address, key []byte) ([]byte, error) {
	res := new(rpc.ResultGetStorage)
	_, err := client.Call(tm.GetStorage, pmap("address", address, "key", key), res)
	if err != nil {
		return nil, err
	}
	return res.Value, nil
}

func CallCode(client RPCClient, fromAddress crypto.Address, code, data []byte) (*rpc.ResultCall, error) {
	res := new(rpc.ResultCall)
	_, err := client.Call(tm.CallCode, pmap("fromAddress", fromAddress, "code", code, "data", data), res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func Call(client RPCClient, fromAddress, toAddress crypto.Address, data []byte) (*rpc.ResultCall, error) {
	res := new(rpc.ResultCall)
	_, err := client.Call(tm.Call, pmap("fromAddress", fromAddress, "toAddress", toAddress,
		"data", data), res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func GetName(client RPCClient, name string) (*names.NameRegEntry, error) {
	res := new(rpc.ResultGetName)
	_, err := client.Call(tm.GetName, pmap("name", name), res)
	if err != nil {
		return nil, err
	}
	return res.Entry, nil
}

func ListBlocks(client RPCClient, minHeight, maxHeight int) (*rpc.ResultListBlocks, error) {
	res := new(rpc.ResultListBlocks)
	_, err := client.Call(tm.ListBlocks, pmap("minHeight", minHeight, "maxHeight", maxHeight), res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func GetBlock(client RPCClient, height int) (*rpc.ResultGetBlock, error) {
	res := new(rpc.ResultGetBlock)
	_, err := client.Call(tm.GetBlock, pmap("height", height), res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func ListUnconfirmedTxs(client RPCClient, maxTxs int) (*rpc.ResultListUnconfirmedTxs, error) {
	res := new(rpc.ResultListUnconfirmedTxs)
	_, err := client.Call(tm.ListUnconfirmedTxs, pmap("maxTxs", maxTxs), res)
	if err != nil {
		return nil, err
	}
	resCon := res
	return resCon, nil
}

func ListValidators(client RPCClient) (*rpc.ResultListValidators, error) {
	res := new(rpc.ResultListValidators)
	_, err := client.Call(tm.ListValidators, pmap(), res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func DumpConsensusState(client RPCClient) (*rpc.ResultDumpConsensusState, error) {
	res := new(rpc.ResultDumpConsensusState)
	_, err := client.Call(tm.DumpConsensusState, pmap(), res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func pmap(keyvals ...interface{}) map[string]interface{} {
	pm, err := paramsMap(keyvals...)
	if err != nil {
		panic(err)
	}
	return pm
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
