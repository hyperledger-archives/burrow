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

func Status(client RPCClient) (*rpc.ResultStatus, error) {
	res := new(rpc.Result)
	_, err := client.Call(method.Status, pmap(), res)
	if err != nil {
		return nil, err
	}
	return res.Unwrap().(*rpc.ResultStatus), nil
}

func ChainId(client RPCClient) (*rpc.ResultChainId, error) {
	res := new(rpc.Result)
	_, err := client.Call(method.ChainID, pmap(), &res)
	if err != nil {
		return nil, err
	}
	return res.Unwrap().(*rpc.ResultChainId), nil
}

func GenPrivAccount(client RPCClient) (*rpc.ResultGeneratePrivateAccount, error) {
	res := new(rpc.Result)
	_, err := client.Call(method.GeneratePrivateAccount, pmap(), res)
	if err != nil {
		return nil, err
	}
	return res.Unwrap().(*rpc.ResultGeneratePrivateAccount), nil
}

func GetAccount(client RPCClient, address acm.Address) (acm.Account, error) {
	res := new(rpc.Result)
	_, err := client.Call(method.GetAccount, pmap("address", address), res)
	if err != nil {
		return nil, err
	}
	return res.Unwrap().(*rpc.ResultGetAccount).Account.Account(), nil
}

func SignTx(client RPCClient, tx txs.Tx, privAccounts []*acm.ConcretePrivateAccount) (txs.Tx, error) {
	res := new(rpc.Result)
	_, err := client.Call(method.SignTx, pmap("tx", tx, "privAccounts", privAccounts), res)
	if err != nil {
		return nil, err
	}
	return res.Unwrap().(*rpc.ResultSignTx).Tx, nil
}

func BroadcastTx(client RPCClient, tx txs.Tx) (*txs.Receipt, error) {
	res := new(rpc.Result)
	_, err := client.Call(method.BroadcastTx, pmap("tx", txs.Wrap(tx)), res)
	if err != nil {
		return nil, err
	}
	return res.Unwrap().(*rpc.ResultBroadcastTx).Receipt, nil
}

func DumpStorage(client RPCClient, address acm.Address) (*rpc.ResultDumpStorage, error) {
	res := new(rpc.Result)
	_, err := client.Call(method.DumpStorage, pmap("address", address), res)
	if err != nil {
		return nil, err
	}
	return res.Unwrap().(*rpc.ResultDumpStorage), nil
}

func GetStorage(client RPCClient, address acm.Address, key []byte) ([]byte, error) {
	res := new(rpc.Result)
	_, err := client.Call(method.GetStorage, pmap("address", address, "key", key), res)
	if err != nil {
		return nil, err
	}
	return res.Unwrap().(*rpc.ResultGetStorage).Value, nil
}

func CallCode(client RPCClient, fromAddress acm.Address, code, data []byte) (*rpc.ResultCall, error) {
	res := new(rpc.Result)
	_, err := client.Call(method.CallCode, pmap("fromAddress", fromAddress, "code", code, "data", data), res)
	if err != nil {
		return nil, err
	}
	return res.Unwrap().(*rpc.ResultCall), nil
}

func Call(client RPCClient, fromAddress, toAddress acm.Address, data []byte) (*rpc.ResultCall, error) {
	res := new(rpc.Result)
	_, err := client.Call(method.Call, pmap("fromAddress", fromAddress, "toAddress", toAddress,
		"data", data), res)
	if err != nil {
		return nil, err
	}
	return res.Unwrap().(*rpc.ResultCall), nil
}

func GetName(client RPCClient, name string) (*execution.NameRegEntry, error) {
	res := new(rpc.Result)
	_, err := client.Call(method.GetName, pmap("name", name), res)
	if err != nil {
		return nil, err
	}
	return res.Unwrap().(*rpc.ResultGetName).Entry, nil
}

func BlockchainInfo(client RPCClient, minHeight, maxHeight int) (*rpc.ResultBlockchainInfo, error) {
	res := new(rpc.Result)
	_, err := client.Call(method.Blockchain, pmap("minHeight", minHeight, "maxHeight", maxHeight), res)
	if err != nil {
		return nil, err
	}
	return res.Unwrap().(*rpc.ResultBlockchainInfo), nil
}

func GetBlock(client RPCClient, height int) ( *rpc.ResultGetBlock,  error) {
	res := new(rpc.Result)
	_, err := client.Call(method.GetBlock, pmap("height", height), res)
	if err != nil {
		return nil, err
	}
	return res.Unwrap().(*rpc.ResultGetBlock), nil
}

func ListUnconfirmedTxs(client RPCClient, maxTxs int) (*rpc.ResultListUnconfirmedTxs, error) {
	res := new(rpc.Result)
	_, err := client.Call(method.ListUnconfirmedTxs, pmap("maxTxs", maxTxs), res)
	if err != nil {
		return nil, err
	}
	resCon := res.Unwrap().(*rpc.ResultListUnconfirmedTxs)
	return resCon, nil
}

func ListValidators(client RPCClient) (*rpc.ResultListValidators, error) {
	res := new(rpc.Result)
	_, err := client.Call(method.ListValidators, pmap(), res)
	if err != nil {
		return nil, err
	}
	return res.Unwrap().(*rpc.ResultListValidators), nil
}

func DumpConsensusState(client RPCClient) (*rpc.ResultDumpConsensusState, error) {
	res := new(rpc.Result)
	_, err := client.Call(method.DumpConsensusState, pmap(), res)
	if err != nil {
		return nil, err
	}
	return res.Unwrap().(*rpc.ResultDumpConsensusState), nil
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
