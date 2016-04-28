package core

import (
	rpc "github.com/tendermint/go-rpc/server"
)

// TODO: eliminate redundancy between here and reading code from core/
var Routes = map[string]*rpc.RPCFunc{
	"status":   rpc.NewRPCFunc(Status, ""),
	"net_info": rpc.NewRPCFunc(NetInfo, ""),
	//	"blockchain":              rpc.NewRPCFunc(BlockchainInfo, "minHeight,maxHeight"),
	"genesis": rpc.NewRPCFunc(Genesis, ""),
	//	"get_block":               rpc.NewRPCFunc(GetBlock, "height"),
	"get_account": rpc.NewRPCFunc(GetAccount, "address"),
	"get_storage": rpc.NewRPCFunc(GetStorage, "address,key"),
	"call":        rpc.NewRPCFunc(Call, "fromAddress,toAddress,data"),
	"call_code":   rpc.NewRPCFunc(CallCode, "fromAddress,code,data"),
	//"list_validators":         rpc.NewRPCFunc(ListValidators, ""),
	// "dump_consensus_state":    rpc.NewRPCFunc(DumpConsensusState, ""),
	"dump_storage": rpc.NewRPCFunc(DumpStorage, "address"),
	// "broadcast_tx":            rpc.NewRPCFunc(BroadcastTx, "tx"),
	// "list_unconfirmed_txs":    rpc.NewRPCFunc(ListUnconfirmedTxs, ""),
	"list_accounts":           rpc.NewRPCFunc(ListAccounts, ""),
	"get_name":                rpc.NewRPCFunc(GetName, "name"),
	"list_names":              rpc.NewRPCFunc(ListNames, ""),
	"unsafe/gen_priv_account": rpc.NewRPCFunc(GenPrivAccount, ""),
	"unsafe/sign_tx":          rpc.NewRPCFunc(SignTx, "tx,privAccounts"),
	// subscribe/unsubscribe are reserved for websocket events.
}
