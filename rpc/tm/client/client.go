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
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/rpc/tm/method"
	"github.com/hyperledger/burrow/txs"
)

type RPCClient interface {
	Call(method string, params map[string]interface{}, result interface{}) (interface{}, error)
}

func Status(client RPCClient) (res *rpc.ResultStatus, err error) {
	_, err = client.Call(method.Status, pmap(), &res)
	return
}

func ChainId(client RPCClient) (res *rpc.ResultChainId, err error) {
	_, err = client.Call(method.ChainID, pmap(), &res)
	return
}

func GenPrivAccount(client RPCClient) (res *rpc.ResultGeneratePrivateAccount, err error) {
	_, err = client.Call(method.GeneratePrivateAccount, pmap(), res)
	return
}

func GetAccount(client RPCClient, address acm.Address) (acm.Account, error) {
	res := new(rpc.ResultGetAccount)
	_, err := client.Call(method.GetAccount, pmap("address", address), res)
	if err != nil {
		return nil, err
	}
	return res.Account.Account(), nil
}

func SignTx(client RPCClient, tx txs.Tx, privAccounts []*acm.ConcretePrivateAccount) (txs.Tx, error) {
	res := new(rpc.ResultSignTx)
	_, err := client.Call(method.SignTx, pmap("tx", tx, "privAccounts", privAccounts), res)
	if err != nil {
		return nil, err
	}
	return res.Tx, nil
}

func BroadcastTx(client RPCClient, tx txs.Tx) (*txs.Receipt, error) {
	res := new(rpc.ResultBroadcastTx)
	_, err := client.Call(method.BroadcastTx, pmap("tx", txs.Wrap(tx)), res)
	if err != nil {
		return nil, err
	}
	return res.Receipt, nil
}

func DumpStorage(client RPCClient, address acm.Address) (res *rpc.ResultDumpStorage, err error) {
	_, err = client.Call(method.DumpStorage, pmap("address", address), res)
	return
}

func GetStorage(client RPCClient, address acm.Address, key []byte) ([]byte, error) {
	res := new(rpc.ResultGetStorage)
	_, err := client.Call(method.GetStorage, pmap("address", address, "key", key), res)
	if err != nil {
		return nil, err
	}
	return res.Value, nil
}

func CallCode(client RPCClient, fromAddress, code, data []byte) (res *rpc.ResultCall, err error) {
	_, err = client.Call(method.CallCode, pmap("fromAddress", fromAddress, "code", code, "data", data), res)
	if err != nil {
		return nil, err
	}
	return
}

func Call(client RPCClient, fromAddress, toAddress acm.Address, data []byte) (res *rpc.ResultCall, err error) {
	_, err = client.Call(method.Call, pmap("fromAddress", fromAddress, "toAddress", toAddress,
		"data", data), res)
	if err != nil {
		return nil, err
	}
	return
}

func GetName(client RPCClient, name string) (*execution.NameRegEntry, error) {
	res := new(rpc.ResultGetName)
	_, err := client.Call(method.GetName, pmap("name", name), res)
	if err != nil {
		return nil, err
	}
	return res.Entry, nil
}

func BlockchainInfo(client RPCClient, minHeight, maxHeight int) (res *rpc.ResultBlockchainInfo, err error) {
	_, err = client.Call(method.Blockchain, pmap("minHeight", minHeight, "maxHeight", maxHeight), res)
	if err != nil {
		return nil, err
	}
	return
}

func GetBlock(client RPCClient, height int) (res *rpc.ResultGetBlock, err error) {
	_, err = client.Call(method.GetBlock, pmap("height", height), res)
	if err != nil {
		return nil, err
	}
	return
}

func ListUnconfirmedTxs(client RPCClient) (res *rpc.ResultListUnconfirmedTxs, err error) {
	_, err = client.Call(method.ListUnconfirmedTxs, pmap(), res)
	return
}

func ListValidators(client RPCClient) (res *rpc.ResultListValidators, err error) {
	_, err = client.Call(method.ListValidators, pmap(), res)
	return
}

func DumpConsensusState(client RPCClient) (res *rpc.ResultDumpConsensusState, err error) {
	_, err = client.Call(method.DumpConsensusState, pmap(), res)
	return
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
