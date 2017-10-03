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

package tm

import (
	"github.com/hyperledger/burrow/rpc/core"
	rpc "github.com/tendermint/tendermint/rpc/lib/server"
)

func GetRoutes(service core.Service) map[string]*rpc.RPCFunc {
	return map[string]*rpc.RPCFunc{
		// Events
		"subscribe":   rpc.NewWSRPCFunc(service.Subscribe, "eventId"),
		"unsubscribe": rpc.NewWSRPCFunc(service.Unsubscribe, "subscriptionId"),

		// Status
		"status":   rpc.NewRPCFunc(service.Status, ""),
		"net_info": rpc.NewRPCFunc(service.NetInfo, ""),

		// Accounts
		"list_accounts": rpc.NewRPCFunc(service.ListAccounts, ""),
		"get_account":   rpc.NewRPCFunc(service.GetAccount, "address"),
		"get_storage":   rpc.NewRPCFunc(service.GetStorage, "address,key"),
		"dump_storage":  rpc.NewRPCFunc(service.DumpStorage, "address"),

		// Simulated call
		"call":      rpc.NewRPCFunc(service.Call, "fromAddress,toAddress,data"),
		"call_code": rpc.NewRPCFunc(service.CallCode, "fromAddress,code,data"),

		// Names
		"get_name":     rpc.NewRPCFunc(service.GetName, "name"),
		"list_names":   rpc.NewRPCFunc(service.ListNames, ""),
		"broadcast_tx": rpc.NewRPCFunc(service.BroadcastTx, "tx"),

		// Blockchain
		"genesis":    rpc.NewRPCFunc(service.Genesis, ""),
		"chain_id":   rpc.NewRPCFunc(service.ChainId, ""),
		"blockchain": rpc.NewRPCFunc(service.BlockchainInfo, "minHeight,maxHeight"),
		"get_block":  rpc.NewRPCFunc(service.GetBlock, "height"),

		// Consensus
		"list_unconfirmed_txs": rpc.NewRPCFunc(service.ListUnconfirmedTxs, ""),
		"list_validators":      rpc.NewRPCFunc(service.ListValidators, ""),
		"dump_consensus_state": rpc.NewRPCFunc(service.DumpConsensusState, ""),

		// Private keys and signing
		"unsafe/gen_priv_account": rpc.NewRPCFunc(service.GeneratePrivateAccount, ""),
		"unsafe/sign_tx":          rpc.NewRPCFunc(service.SignTx, "tx,privAccounts"),
	}
}
