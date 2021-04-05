// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

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

func Status(client rpc.Client) (*rpc.ResultStatus, error) {
	res := new(rpc.ResultStatus)
	err := client.Call(rpcinfo.Status, pmap(), res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func ChainId(client rpc.Client) (*rpc.ResultChainId, error) {
	res := new(rpc.ResultChainId)
	err := client.Call(rpcinfo.ChainID, pmap(), &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func Account(client rpc.Client, address crypto.Address) (*acm.Account, error) {
	res := new(rpc.ResultAccount)
	err := client.Call(rpcinfo.Account, pmap("address", address), res)
	if err != nil {
		return nil, err
	}
	return res.Account, nil
}

func DumpStorage(client rpc.Client, address crypto.Address) (*rpc.ResultDumpStorage, error) {
	res := new(rpc.ResultDumpStorage)
	err := client.Call(rpcinfo.DumpStorage, pmap("address", address), res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func Storage(client rpc.Client, address crypto.Address, key []byte) ([]byte, error) {
	res := new(rpc.ResultStorage)
	err := client.Call(rpcinfo.Storage, pmap("address", address, "key", key), res)
	if err != nil {
		return nil, err
	}
	return res.Value, nil
}

func Name(client rpc.Client, name string) (*names.Entry, error) {
	res := new(rpc.ResultName)
	err := client.Call(rpcinfo.Name, pmap("name", name), res)
	if err != nil {
		return nil, err
	}
	return res.Entry, nil
}

func Names(client rpc.Client, regex string) ([]*names.Entry, error) {
	res := new(rpc.ResultNames)
	err := client.Call(rpcinfo.Names, pmap("regex", regex), res)
	if err != nil {
		return nil, err
	}
	return res.Names, nil
}

func Blocks(client rpc.Client, minHeight, maxHeight int) (*rpc.ResultBlocks, error) {
	res := new(rpc.ResultBlocks)
	err := client.Call(rpcinfo.Blocks, pmap("minHeight", minHeight, "maxHeight", maxHeight), res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func Block(client rpc.Client, height int) (*rpc.ResultBlock, error) {
	res := new(rpc.ResultBlock)
	err := client.Call(rpcinfo.Block, pmap("height", height), res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func UnconfirmedTxs(client rpc.Client, maxTxs int) (*rpc.ResultUnconfirmedTxs, error) {
	res := new(rpc.ResultUnconfirmedTxs)
	err := client.Call(rpcinfo.UnconfirmedTxs, pmap("maxTxs", maxTxs), res)
	if err != nil {
		return nil, err
	}
	resCon := res
	return resCon, nil
}

func Validators(client rpc.Client) (*rpc.ResultValidators, error) {
	res := new(rpc.ResultValidators)
	err := client.Call(rpcinfo.Validators, pmap(), res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func Consensus(client rpc.Client) (*rpc.ResultConsensusState, error) {
	res := new(rpc.ResultConsensusState)
	err := client.Call(rpcinfo.Consensus, pmap(), res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func pmap(keyvals ...interface{}) map[string]interface{} {
	pm, err := ParamsToMap(keyvals...)
	if err != nil {
		panic(err)
	}
	return pm
}

func ParamsToMap(orderedKeyVals ...interface{}) (map[string]interface{}, error) {
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
