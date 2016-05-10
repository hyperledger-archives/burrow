package client

import (
	rpcclient "github.com/tendermint/go-rpc/client"

	acm "github.com/eris-ltd/eris-db/account"
	ctypes "github.com/eris-ltd/eris-db/rpc/core/types"
	"github.com/eris-ltd/eris-db/txs"
)

func Status(client rpcclient.Client) (*ctypes.ResultStatus, error) {
	var res ctypes.ErisDBResult //ResultStatus)
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
	return res.(*ctypes.ResultStatus), nil
}

func GenPrivAccount(client rpcclient.Client) (*acm.PrivAccount, error) {
	var res ctypes.ErisDBResult
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
	return res.(*ctypes.ResultGenPrivAccount).PrivAccount, nil
}

func GetAccount(client rpcclient.Client, addr []byte) (*acm.Account, error) {
	var res ctypes.ErisDBResult
	var err error
	switch cli := client.(type) {
	case *rpcclient.ClientJSONRPC:
		_, err = cli.Call("get_account", []interface{}{addr}, &res)
	case *rpcclient.ClientURI:
		_, err = cli.Call("get_account", map[string]interface{}{"address": addr}, &res)
	}
	if err != nil {
		return nil, err
	}
	return res.(*ctypes.ResultGetAccount).Account, nil
}

func SignTx(client rpcclient.Client, tx types.Tx, privAccs []*acm.PrivAccount) (types.Tx, error) {
	wrapTx := struct {
		types.Tx `json:"unwrap"`
	}{tx}
	var res ctypes.ErisDBResult
	var err error
	switch cli := client.(type) {
	case *rpcclient.ClientJSONRPC:
		_, err = cli.Call("unsafe/sign_tx", []interface{}{wrapTx, privAccs}, &res)
	case *rpcclient.ClientURI:
		_, err = cli.Call("unsafe/sign_tx", map[string]interface{}{"tx": wrapTx, "privAccounts": privAccs}, &res)
	}
	if err != nil {
		return nil, err
	}
	return res.(*ctypes.ResultSignTx).Tx, nil
}

func BroadcastTx(client rpcclient.Client, tx types.Tx) (ctypes.Receipt, error) {
	wrapTx := struct {
		types.Tx `json:"unwrap"`
	}{tx}
	var res ctypes.ErisDBResult
	var err error
	switch cli := client.(type) {
	case *rpcclient.ClientJSONRPC:
		_, err = cli.Call("broadcast_tx", []interface{}{wrapTx}, &res)
	case *rpcclient.ClientURI:
		_, err = cli.Call("broadcast_tx", map[string]interface{}{"tx": wrapTx}, &res)
	}
	if err != nil {
		return ctypes.Receipt{}, err
	}
	data := res.(*ctypes.ResultBroadcastTx).Data
	// TODO: unmarshal data to receuipt
	_ = data
	return ctypes.Receipt{}, err

}

func DumpStorage(client rpcclient.Client, addr []byte) (*ctypes.ResultDumpStorage, error) {
	var res ctypes.ErisDBResult
	var err error
	switch cli := client.(type) {
	case *rpcclient.ClientJSONRPC:
		_, err = cli.Call("dump_storage", []interface{}{addr}, &res)
	case *rpcclient.ClientURI:
		_, err = cli.Call("dump_storage", map[string]interface{}{"address": addr}, &res)
	}
	if err != nil {
		return nil, err
	}
	return res.(*ctypes.ResultDumpStorage), err
}

func GetStorage(client rpcclient.Client, addr, key []byte) ([]byte, error) {
	var res ctypes.ErisDBResult
	var err error
	switch cli := client.(type) {
	case *rpcclient.ClientJSONRPC:
		_, err = cli.Call("get_storage", []interface{}{addr, key}, &res)
	case *rpcclient.ClientURI:
		_, err = cli.Call("get_storage", map[string]interface{}{"address": addr, "key": key}, &res)
	}
	if err != nil {
		return nil, err
	}
	return res.(*ctypes.ResultGetStorage).Value, nil
}

func CallCode(client rpcclient.Client, fromAddress, code, data []byte) (*ctypes.ResultCall, error) {
	var res ctypes.ErisDBResult
	var err error
	switch cli := client.(type) {
	case *rpcclient.ClientJSONRPC:
		_, err = cli.Call("call_code", []interface{}{fromAddress, code, data}, &res)
	case *rpcclient.ClientURI:
		_, err = cli.Call("call_code", map[string]interface{}{"fromAddress": fromAddress, "code": code, "data": data}, &res)

	}
	if err != nil {
		return nil, err
	}
	return res.(*ctypes.ResultCall), err
}

func Call(client rpcclient.Client, fromAddress, toAddress, data []byte) (*ctypes.ResultCall, error) {
	var res ctypes.ErisDBResult
	var err error
	switch cli := client.(type) {
	case *rpcclient.ClientJSONRPC:
		_, err = cli.Call("call", []interface{}{fromAddress, toAddress, data}, &res)
	case *rpcclient.ClientURI:
		_, err = cli.Call("call", map[string]interface{}{"fromAddress": fromAddress, "toAddress": toAddress, "data": data}, &res)

	}
	if err != nil {
		return nil, err
	}
	return res.(*ctypes.ResultCall), err
}

func GetName(client rpcclient.Client, name string) (*types.NameRegEntry, error) {
	var res ctypes.ErisDBResult
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
	return res.(*ctypes.ResultGetName).Entry, nil
}
