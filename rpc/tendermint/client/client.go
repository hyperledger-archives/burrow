package client

import (
	"fmt"
	acm "github.com/eris-ltd/eris-db/account"
	core_types "github.com/eris-ltd/eris-db/core/types"
	rpc_types "github.com/eris-ltd/eris-db/rpc/tendermint/core/types"
	"github.com/eris-ltd/eris-db/txs"
	rpcclient "github.com/tendermint/go-rpc/client"
	"github.com/tendermint/go-wire"
)

func Status(client rpcclient.Client) (*rpc_types.ResultStatus, error) {
	var res rpc_types.ErisDBResult //ResultStatus)
	var err error
	switch cli := client.(type) {
	case *rpcclient.ClientJSONRPC:
		_, err = cli.Call("status", []interface{}{}, &res)
	case *rpcclient.ClientURI:
		_, err = cli.Call("status", map[string]interface{}{}, &res)
	}
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultStatus), nil
}

func GenPrivAccount(client rpcclient.Client) (*acm.PrivAccount, error) {
	var res rpc_types.ErisDBResult
	var err error
	switch cli := client.(type) {
	case *rpcclient.ClientJSONRPC:
		_, err = cli.Call("unsafe/gen_priv_account", []interface{}{}, &res)
	case *rpcclient.ClientURI:
		_, err = cli.Call("unsafe/gen_priv_account", map[string]interface{}{}, &res)
	}
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultGenPrivAccount).PrivAccount, nil
}

func GetAccount(client rpcclient.Client, addr []byte) (*acm.Account, error) {
	var res rpc_types.ErisDBResult
	var err error
	switch cli := client.(type) {
	case *rpcclient.ClientJSONRPC:
		_, err = cli.Call("get_account", []interface{}{addr}, &res)
	case *rpcclient.ClientURI:
		_, err = cli.Call("get_account", map[string]interface{}{"address": addr},
			&res)
	}
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultGetAccount).Account, nil
}

func SignTx(client rpcclient.Client, tx txs.Tx,
	privAccs []*acm.PrivAccount) (txs.Tx, error) {
	wrapTx := struct {
		txs.Tx `json:"unwrap"`
	}{tx}
	var res rpc_types.ErisDBResult
	var err error
	switch cli := client.(type) {
	case *rpcclient.ClientJSONRPC:
		_, err = cli.Call("unsafe/sign_tx", []interface{}{wrapTx, privAccs}, &res)
	case *rpcclient.ClientURI:
		_, err = cli.Call("unsafe/sign_tx", map[string]interface{}{
			"tx":           wrapTx,
			"privAccounts": privAccs,
		}, &res)
	}
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultSignTx).Tx, nil
}

func BroadcastTx(client rpcclient.Client,
	tx txs.Tx) (txs.Receipt, error) {
	wrapTx := struct {
		txs.Tx `json:"unwrap"`
	}{tx}
	var res rpc_types.ErisDBResult
	var err error
	switch cli := client.(type) {
	case *rpcclient.ClientJSONRPC:
		_, err = cli.Call("broadcast_tx", []interface{}{wrapTx}, &res)
	case *rpcclient.ClientURI:
		_, err = cli.Call("broadcast_tx", map[string]interface{}{"tx": wrapTx},
			&res)
	}
	if err != nil {
		return txs.Receipt{}, err
	}
	receiptBytes := res.(*rpc_types.ResultBroadcastTx).Data
	receipt := txs.Receipt{}
	err = wire.ReadBinaryBytes(receiptBytes, &receipt)
	return receipt, err

}

func DumpStorage(client rpcclient.Client,
	addr []byte) (*rpc_types.ResultDumpStorage, error) {
	var res rpc_types.ErisDBResult
	var err error
	switch cli := client.(type) {
	case *rpcclient.ClientJSONRPC:
		_, err = cli.Call("dump_storage", []interface{}{addr}, &res)
	case *rpcclient.ClientURI:
		_, err = cli.Call("dump_storage", map[string]interface{}{"address": addr},
			&res)
	}
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultDumpStorage), err
}

func GetStorage(client rpcclient.Client, addr, key []byte) ([]byte, error) {
	var res rpc_types.ErisDBResult
	var err error
	switch cli := client.(type) {
	case *rpcclient.ClientJSONRPC:
		_, err = cli.Call("get_storage", []interface{}{addr, key}, &res)
	case *rpcclient.ClientURI:
		_, err = cli.Call("get_storage", map[string]interface{}{
			"address": addr,
			"key":     key,
		}, &res)
	}
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultGetStorage).Value, nil
}

func CallCode(client rpcclient.Client,
	fromAddress, code, data []byte) (*rpc_types.ResultCall, error) {
	var res rpc_types.ErisDBResult
	var err error
	switch cli := client.(type) {
	case *rpcclient.ClientJSONRPC:
		_, err = cli.Call("call_code", []interface{}{fromAddress, code, data}, &res)
	case *rpcclient.ClientURI:
		_, err = cli.Call("call_code", map[string]interface{}{
			"fromAddress": fromAddress,
			"code":        code,
			"data":        data,
		}, &res)

	}
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultCall), err
}

func Call(client rpcclient.Client, fromAddress, toAddress,
	data []byte) (*rpc_types.ResultCall, error) {
	var res rpc_types.ErisDBResult
	var err error
	switch cli := client.(type) {
	case *rpcclient.ClientJSONRPC:
		_, err = cli.Call("call", []interface{}{fromAddress, toAddress, data}, &res)
	case *rpcclient.ClientURI:
		_, err = cli.Call("call", map[string]interface{}{
			"fromAddress": fromAddress,
			"toAddress":   toAddress,
			"data":        data,
		}, &res)

	}
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultCall), err
}

func GetName(client rpcclient.Client, name string) (*core_types.NameRegEntry, error) {
	var res rpc_types.ErisDBResult
	var err error
	switch cli := client.(type) {
	case *rpcclient.ClientJSONRPC:
		_, err = cli.Call("get_name", []interface{}{name}, &res)
	case *rpcclient.ClientURI:
		_, err = cli.Call("get_name", map[string]interface{}{"name": name}, &res)
	}
	if err != nil {
		return nil, err
	}
	return res.(*rpc_types.ResultGetName).Entry, nil
}
