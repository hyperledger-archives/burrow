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

package infoclient

import (
	"errors"
	"fmt"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/rpc/rpcinfo"
)

type RPCClient interface {
	Call(method string, params map[string]interface{}, result interface{}) (interface{}, error)
}

func Status(client RPCClient) (*rpc.ResultStatus, error) {
	res := new(rpc.ResultStatus)
	_, err := client.Call(rpcinfo.Status, pmap(), res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func ChainId(client RPCClient) (*rpc.ResultChainId, error) {
	res := new(rpc.ResultChainId)
	_, err := client.Call(rpcinfo.ChainID, pmap(), &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func Account(client RPCClient, address crypto.Address) (acm.Account, error) {
	res := new(rpc.ResultAccount)
	_, err := client.Call(rpcinfo.Account, pmap("address", address), res)
	if err != nil {
		return nil, err
	}
	concreteAccount := res.Account
	if concreteAccount == nil {
		return nil, nil
	}
	return concreteAccount.Account(), nil
}

func DumpStorage(client RPCClient, address crypto.Address) (*rpc.ResultDumpStorage, error) {
	res := new(rpc.ResultDumpStorage)
	_, err := client.Call(rpcinfo.DumpStorage, pmap("address", address), res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func Storage(client RPCClient, address crypto.Address, key []byte) ([]byte, error) {
	res := new(rpc.ResultStorage)
	_, err := client.Call(rpcinfo.Storage, pmap("address", address, "key", key), res)
	if err != nil {
		return nil, err
	}
	return res.Value, nil
}

func Name(client RPCClient, name string) (*names.Entry, error) {
	res := new(rpc.ResultName)
	_, err := client.Call(rpcinfo.Name, pmap("name", name), res)
	if err != nil {
		return nil, err
	}
	return res.Entry, nil
}

func Blocks(client RPCClient, minHeight, maxHeight int) (*rpc.ResultBlocks, error) {
	res := new(rpc.ResultBlocks)
	_, err := client.Call(rpcinfo.Blocks, pmap("minHeight", minHeight, "maxHeight", maxHeight), res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func Block(client RPCClient, height int) (*rpc.ResultBlock, error) {
	res := new(rpc.ResultBlock)
	_, err := client.Call(rpcinfo.Block, pmap("height", height), res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func UnconfirmedTxs(client RPCClient, maxTxs int) (*rpc.ResultUnconfirmedTxs, error) {
	res := new(rpc.ResultUnconfirmedTxs)
	_, err := client.Call(rpcinfo.UnconfirmedTxs, pmap("maxTxs", maxTxs), res)
	if err != nil {
		return nil, err
	}
	resCon := res
	return resCon, nil
}

func Validators(client RPCClient) (*rpc.ResultValidators, error) {
	res := new(rpc.ResultValidators)
	_, err := client.Call(rpcinfo.Validators, pmap(), res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func Consensus(client RPCClient) (*rpc.ResultConsensusState, error) {
	res := new(rpc.ResultConsensusState)
	_, err := client.Call(rpcinfo.Consensus, pmap(), res)
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
