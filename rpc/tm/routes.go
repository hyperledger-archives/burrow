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
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/rpc/tm/method"
	gorpc "github.com/tendermint/tendermint/rpc/lib/server"
)

func GetRoutes(service rpc.Service) map[string]*gorpc.RPCFunc {
	return map[string]*gorpc.RPCFunc{
		// Events
		method.Subscribe:   gorpc.NewWSRPCFunc(service.Subscribe, "eventId"),
		method.Unsubscribe: gorpc.NewWSRPCFunc(service.Unsubscribe, "subscriptionId"),

		// Status
		method.Status:  gorpc.NewRPCFunc(service.Status, ""),
		method.NetInfo: gorpc.NewRPCFunc(service.NetInfo, ""),

		// Accounts
		method.ListAccounts: gorpc.NewRPCFunc(service.ListAccounts, ""),
		method.GetAccount:   gorpc.NewRPCFunc(service.GetAccount, "address"),
		method.GetStorage:   gorpc.NewRPCFunc(service.GetStorage, "address,key"),
		method.DumpStorage:  gorpc.NewRPCFunc(service.DumpStorage, "address"),

		// Simulated call
		method.Call:     gorpc.NewRPCFunc(service.Call, "fromAddress,toAddress,data"),
		method.CallCode: gorpc.NewRPCFunc(service.CallCode, "fromAddress,code,data"),

		// Names
		method.GetName:     gorpc.NewRPCFunc(service.GetName, "name"),
		method.ListNames:   gorpc.NewRPCFunc(service.ListNames, ""),
		method.BroadcastTx: gorpc.NewRPCFunc(service.BroadcastTx, "tx"),

		// Blockchain
		method.Genesis:    gorpc.NewRPCFunc(service.Genesis, ""),
		method.ChainID:    gorpc.NewRPCFunc(service.ChainId, ""),
		method.Blockchain: gorpc.NewRPCFunc(service.BlockchainInfo, "minHeight,maxHeight"),
		method.GetBlock:   gorpc.NewRPCFunc(service.GetBlock, "height"),

		// Consensus
		method.ListUnconfirmedTxs: gorpc.NewRPCFunc(service.ListUnconfirmedTxs, ""),
		method.ListValidators:     gorpc.NewRPCFunc(service.ListValidators, ""),
		method.DumpConsensusState: gorpc.NewRPCFunc(service.DumpConsensusState, ""),

		// Private keys and signing
		method.GeneratePrivateAccount: gorpc.NewRPCFunc(service.GeneratePrivateAccount, ""),
		method.SignTx:                 gorpc.NewRPCFunc(service.SignTx, "tx,privAccounts"),
	}
}
