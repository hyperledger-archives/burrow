package core

import (
	"fmt"

	acm "github.com/eris-ltd/eris-db/account"
	definitions "github.com/eris-ltd/eris-db/definitions"
	ctypes "github.com/eris-ltd/eris-db/rpc/tendermint/core/types"
	"github.com/eris-ltd/eris-db/txs"
	rpc "github.com/tendermint/go-rpc/server"
)

// TODO: [ben] encapsulate Routes into a struct for a given TendermintPipe

// TODO: eliminate redundancy between here and reading code from core/
// var Routes = map[string]*rpc.RPCFunc{
// 	"status":                  rpc.NewRPCFunc(StatusResult, ""),
// 	"net_info":                rpc.NewRPCFunc(NetInfoResult, ""),
// 	"genesis":                 rpc.NewRPCFunc(GenesisResult, ""),
// 	"get_account":             rpc.NewRPCFunc(GetAccountResult, "address"),
// 	"get_storage":             rpc.NewRPCFunc(GetStorageResult, "address,key"),
// 	"call":                    rpc.NewRPCFunc(CallResult, "fromAddress,toAddress,data"),
// 	"call_code":               rpc.NewRPCFunc(CallCodeResult, "fromAddress,code,data"),
// 	"dump_storage":            rpc.NewRPCFunc(DumpStorageResult, "address"),
// 	"list_accounts":           rpc.NewRPCFunc(ListAccountsResult, ""),
// 	"get_name":                rpc.NewRPCFunc(GetNameResult, "name"),
// 	"list_names":              rpc.NewRPCFunc(ListNamesResult, ""),
// 	"broadcast_tx":            rpc.NewRPCFunc(BroadcastTxResult, "tx"),
// 	"unsafe/gen_priv_account": rpc.NewRPCFunc(GenPrivAccountResult, ""),
// 	"unsafe/sign_tx":          rpc.NewRPCFunc(SignTxResult, "tx,privAccounts"),
//
// 	// TODO: hookup
// 	//	"blockchain":              rpc.NewRPCFunc(BlockchainInfo, "minHeight,maxHeight"),
// 	//	"get_block":               rpc.NewRPCFunc(GetBlock, "height"),
// 	//"list_validators":         rpc.NewRPCFunc(ListValidators, ""),
// 	// "dump_consensus_state":    rpc.NewRPCFunc(DumpConsensusState, ""),
// 	// "list_unconfirmed_txs":    rpc.NewRPCFunc(ListUnconfirmedTxs, ""),
// 	// subscribe/unsubscribe are reserved for websocket events.
// }

type TendermintRoutes struct {
	tendermintPipe definitions.TendermintPipe
}

func (this *TendermintRoutes) GetRoutes() map[string]*rpc.RPCFunc {
	var routes = map[string]*rpc.RPCFunc{
		"status":                  rpc.NewRPCFunc(this.StatusResult, ""),
		"net_info":                rpc.NewRPCFunc(this.NetInfoResult, ""),
		"genesis":                 rpc.NewRPCFunc(this.GenesisResult, ""),
		"get_account":             rpc.NewRPCFunc(this.GetAccountResult, "address"),
		"get_storage":             rpc.NewRPCFunc(this.GetStorageResult, "address,key"),
		"call":                    rpc.NewRPCFunc(this.CallResult, "fromAddress,toAddress,data"),
		"call_code":               rpc.NewRPCFunc(this.CallCodeResult, "fromAddress,code,data"),
		"dump_storage":            rpc.NewRPCFunc(this.DumpStorageResult, "address"),
		"list_accounts":           rpc.NewRPCFunc(this.ListAccountsResult, ""),
		"get_name":                rpc.NewRPCFunc(this.GetNameResult, "name"),
		"list_names":              rpc.NewRPCFunc(this.ListNamesResult, ""),
		"broadcast_tx":            rpc.NewRPCFunc(this.BroadcastTxResult, "tx"),
		"unsafe/gen_priv_account": rpc.NewRPCFunc(this.GenPrivAccountResult, ""),
		"unsafe/sign_tx":          rpc.NewRPCFunc(this.SignTxResult, "tx,privAccounts"),

		// TODO: hookup
		//	"blockchain":              rpc.NewRPCFunc(BlockchainInfo, "minHeight,maxHeight"),
		//	"get_block":               rpc.NewRPCFunc(GetBlock, "height"),
		//"list_validators":         rpc.NewRPCFunc(ListValidators, ""),
		// "dump_consensus_state":    rpc.NewRPCFunc(DumpConsensusState, ""),
		// "list_unconfirmed_txs":    rpc.NewRPCFunc(ListUnconfirmedTxs, ""),
		// subscribe/unsubscribe are reserved for websocket events.
	}
	return routes
}

func (this *TendermintRoutes) StatusResult() (ctypes.ErisDBResult, error) {
	if r, err := this.tendermintPipe.Status(); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func (this *TendermintRoutes) NetInfoResult() (ctypes.ErisDBResult, error) {
	if r, err := this.tendermintPipe.NetInfo(); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func (this *TendermintRoutes) GenesisResult() (ctypes.ErisDBResult, error) {
	if r, err := this.tendermintPipe.Genesis(); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func (this *TendermintRoutes) GetAccountResult(address []byte) (ctypes.ErisDBResult, error) {
	if r, err := this.tendermintPipe.GetAccount(address); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func (this *TendermintRoutes) GetStorageResult(address, key []byte) (ctypes.ErisDBResult, error) {
	if r, err := this.tendermintPipe.GetStorage(address, key); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func (this *TendermintRoutes) CallResult(fromAddress, toAddress, data []byte) (ctypes.ErisDBResult, error) {
	if r, err := this.tendermintPipe.Call(fromAddress, toAddress, data); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func (this *TendermintRoutes) CallCodeResult(fromAddress, code, data []byte) (ctypes.ErisDBResult, error) {
	if r, err := this.tendermintPipe.CallCode(fromAddress, code, data); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func (this *TendermintRoutes) DumpStorageResult(address []byte) (ctypes.ErisDBResult, error) {
	if r, err := this.tendermintPipe.DumpStorage(address); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func (this *TendermintRoutes) ListAccountsResult() (ctypes.ErisDBResult, error) {
	if r, err := this.tendermintPipe.ListAccounts(); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func (this *TendermintRoutes) GetNameResult(name string) (ctypes.ErisDBResult, error) {
	if r, err := this.tendermintPipe.GetName(name); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func (this *TendermintRoutes) ListNamesResult() (ctypes.ErisDBResult, error) {
	if r, err := this.tendermintPipe.ListNames(); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func (this *TendermintRoutes) GenPrivAccountResult() (ctypes.ErisDBResult, error) {
	// if r, err := this.tendermintPipe.GenPrivAccount(); err != nil {
	// 	return nil, err
	// } else {
	// 	return r, nil
	// }
	return nil, fmt.Errorf("Unimplemented as poor practice to generate private account over unencrypted RPC")
}

func (this *TendermintRoutes) SignTxResult(tx txs.Tx, privAccounts []*acm.PrivAccount) (ctypes.ErisDBResult, error) {
	// if r, err := this.tendermintPipe.SignTx(tx, privAccounts); err != nil {
	// 	return nil, err
	// } else {
	// 	return r, nil
	// }
	return nil, fmt.Errorf("Unimplemented as poor practice to pass private account over unencrypted RPC")
}

func (this *TendermintRoutes) BroadcastTxResult(tx txs.Tx) (ctypes.ErisDBResult, error) {
	if r, err := this.tendermintPipe.BroadcastTxSync(tx); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}
