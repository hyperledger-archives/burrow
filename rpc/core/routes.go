package core

import (
	acm "github.com/eris-ltd/eris-db/account"
	ctypes "github.com/eris-ltd/eris-db/rpc/core/types"
	"github.com/eris-ltd/eris-db/txs"
	rpc "github.com/tendermint/go-rpc/server"
)

// TODO: eliminate redundancy between here and reading code from core/
var Routes = map[string]*rpc.RPCFunc{
	"status":                  rpc.NewRPCFunc(StatusResult, ""),
	"net_info":                rpc.NewRPCFunc(NetInfoResult, ""),
	"genesis":                 rpc.NewRPCFunc(GenesisResult, ""),
	"get_account":             rpc.NewRPCFunc(GetAccountResult, "address"),
	"get_storage":             rpc.NewRPCFunc(GetStorageResult, "address,key"),
	"call":                    rpc.NewRPCFunc(CallResult, "fromAddress,toAddress,data"),
	"call_code":               rpc.NewRPCFunc(CallCodeResult, "fromAddress,code,data"),
	"dump_storage":            rpc.NewRPCFunc(DumpStorageResult, "address"),
	"list_accounts":           rpc.NewRPCFunc(ListAccountsResult, ""),
	"get_name":                rpc.NewRPCFunc(GetNameResult, "name"),
	"list_names":              rpc.NewRPCFunc(ListNamesResult, ""),
	"broadcast_tx":            rpc.NewRPCFunc(BroadcastTxResult, "tx"),
	"unsafe/gen_priv_account": rpc.NewRPCFunc(GenPrivAccountResult, ""),
	"unsafe/sign_tx":          rpc.NewRPCFunc(SignTxResult, "tx,privAccounts"),

	// TODO: hookup
	//	"blockchain":              rpc.NewRPCFunc(BlockchainInfo, "minHeight,maxHeight"),
	//	"get_block":               rpc.NewRPCFunc(GetBlock, "height"),
	//"list_validators":         rpc.NewRPCFunc(ListValidators, ""),
	// "dump_consensus_state":    rpc.NewRPCFunc(DumpConsensusState, ""),
	// "list_unconfirmed_txs":    rpc.NewRPCFunc(ListUnconfirmedTxs, ""),
	// subscribe/unsubscribe are reserved for websocket events.
}

func StatusResult() (ctypes.ErisDBResult, error) {
	if r, err := Status(); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func NetInfoResult() (ctypes.ErisDBResult, error) {
	if r, err := NetInfo(); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func GenesisResult() (ctypes.ErisDBResult, error) {
	if r, err := Genesis(); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func GetAccountResult(address []byte) (ctypes.ErisDBResult, error) {
	if r, err := GetAccount(address); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func GetStorageResult(address, key []byte) (ctypes.ErisDBResult, error) {
	if r, err := GetStorage(address, key); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func CallResult(fromAddress, toAddress, data []byte) (ctypes.ErisDBResult, error) {
	if r, err := Call(fromAddress, toAddress, data); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func CallCodeResult(fromAddress, code, data []byte) (ctypes.ErisDBResult, error) {
	if r, err := CallCode(fromAddress, code, data); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func DumpStorageResult(address []byte) (ctypes.ErisDBResult, error) {
	if r, err := DumpStorage(address); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func ListAccountsResult() (ctypes.ErisDBResult, error) {
	if r, err := ListAccounts(); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func GetNameResult(name string) (ctypes.ErisDBResult, error) {
	if r, err := GetName(name); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func ListNamesResult() (ctypes.ErisDBResult, error) {
	if r, err := ListNames(); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func GenPrivAccountResult() (ctypes.ErisDBResult, error) {
	if r, err := GenPrivAccount(); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func SignTxResult(tx types.Tx, privAccounts []*acm.PrivAccount) (ctypes.ErisDBResult, error) {
	if r, err := SignTx(tx, privAccounts); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func BroadcastTxResult(tx types.Tx) (ctypes.ErisDBResult, error) {
	if r, err := BroadcastTxSync(tx); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}
