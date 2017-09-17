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
	"fmt"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/txs"
	rpc "github.com/tendermint/tendermint/rpc/lib/server"
	"github.com/tendermint/tendermint/rpc/lib/types"
)

// Magic! Should probably be configurable, but not shouldn't be so huge we
// end up DoSing ourselves.
const maxBlockLookback = 20

func GetRoutes() map[string]*rpc.RPCFunc {
	return map[string]*rpc.RPCFunc{
		// Events
		"subscribe":               rpc.NewWSRPCFunc(tmRoutes.Subscribe, "eventId"),
		"unsubscribe":             rpc.NewWSRPCFunc(tmRoutes.Unsubscribe, "subscriptionId"),

		// Status
		"status":                  rpc.NewRPCFunc(tmRoutes.StatusResult, ""),
		"net_info":                rpc.NewRPCFunc(tmRoutes.NetInfoResult, ""),

		// Accounts
		"list_accounts":           rpc.NewRPCFunc(tmRoutes.ListAccountsResult, ""),
		"get_account":             rpc.NewRPCFunc(tmRoutes.GetAccountResult, "address"),
		"get_storage":             rpc.NewRPCFunc(tmRoutes.GetStorageResult, "address,key"),
		"dump_storage":            rpc.NewRPCFunc(tmRoutes.DumpStorageResult, "address"),

		// Simulated call
		"call":                    rpc.NewRPCFunc(tmRoutes.CallResult, "fromAddress,toAddress,data"),
		"call_code":               rpc.NewRPCFunc(tmRoutes.CallCodeResult, "fromAddress,code,data"),

		// Names
		"get_name":                rpc.NewRPCFunc(tmRoutes.GetNameResult, "name"),
		"list_names":              rpc.NewRPCFunc(tmRoutes.ListNamesResult, ""),
		"broadcast_tx":            rpc.NewRPCFunc(tmRoutes.BroadcastTxResult, "tx"),

		// Blockchain
		"genesis":                 rpc.NewRPCFunc(tmRoutes.GenesisResult, ""),
		"chain_id":                rpc.NewRPCFunc(tmRoutes.ChainIdResult, ""),
		"blockchain":              rpc.NewRPCFunc(tmRoutes.BlockchainInfo, "minHeight,maxHeight"),
		"get_block":               rpc.NewRPCFunc(tmRoutes.GetBlock, "height"),

		// Consensus
		"list_unconfirmed_txs":    rpc.NewRPCFunc(tmRoutes.ListUnconfirmedTxs, ""),
		"list_validators":         rpc.NewRPCFunc(tmRoutes.ListValidators, ""),
		"dump_consensus_state":    rpc.NewRPCFunc(tmRoutes.DumpConsensusState, ""),

		// Private keys and signing
		"unsafe/gen_priv_account": rpc.NewRPCFunc(tmRoutes.GenPrivAccountResult, ""),
		"unsafe/sign_tx":          rpc.NewRPCFunc(tmRoutes.SignTxResult, "tx,privAccounts"),
		// TODO: [Silas] do we also carry forward "consensus_state" as in v0?
	}
}

func (tmRoutes *TendermintRoutes) Subscribe(wsCtx rpctypes.WSRPCContext,
	eventId string) (ctypes.BurrowResult, error) {
	// NOTE: RPCResponses of subscribed events have id suffix "#event"
	// TODO: we really ought to allow multiple subscriptions from the same client address
	// to the same event. The code as it stands reflects the somewhat broken tendermint
	// implementation. We can use GenerateSubId to randomize the subscriptions id
	// and return it in the result. This would require clients to hang on to a
	// subscription id if they wish to unsubscribe, but then again they can just
	// drop their connection
	result, err := tmRoutes.tendermintPipe.Subscribe(eventId,
		func(result ctypes.BurrowResult) {
			wsCtx.GetRemoteAddr()
			// NOTE: EventSwitch callbacks must be nonblocking
			wsCtx.TryWriteRPCResponse(
				rpctypes.NewRPCResponse(wsCtx.Request.ID+"#event", &result, ""))
		})
	if err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

func (tmRoutes *TendermintRoutes) Unsubscribe(wsCtx rpctypes.WSRPCContext,
	subscriptionId string) (ctypes.BurrowResult, error) {
	result, err := tmRoutes.tendermintPipe.Unsubscribe(subscriptionId)
	if err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

func (tmRoutes *TendermintRoutes) StatusResult() (ctypes.BurrowResult, error) {
	if r, err := tmRoutes.tendermintPipe.Status(); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func (tmRoutes *TendermintRoutes) NetInfoResult() (ctypes.BurrowResult, error) {
	if r, err := tmRoutes.tendermintPipe.NetInfo(); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func (tmRoutes *TendermintRoutes) GenesisResult() (ctypes.BurrowResult, error) {
	if r, err := tmRoutes.tendermintPipe.Genesis(); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func (tmRoutes *TendermintRoutes) ChainIdResult() (ctypes.BurrowResult, error) {
	if r, err := tmRoutes.tendermintPipe.ChainId(); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func (tmRoutes *TendermintRoutes) GetAccountResult(address []byte) (ctypes.BurrowResult, error) {
	if r, err := tmRoutes.tendermintPipe.GetAccount(address); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func (tmRoutes *TendermintRoutes) GetStorageResult(address, key []byte) (ctypes.BurrowResult, error) {
	if r, err := tmRoutes.tendermintPipe.GetStorage(address, key); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func (tmRoutes *TendermintRoutes) CallResult(fromAddress, toAddress,
	data []byte) (ctypes.BurrowResult, error) {
	if r, err := tmRoutes.tendermintPipe.Call(fromAddress, toAddress, data); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func (tmRoutes *TendermintRoutes) CallCodeResult(fromAddress, code,
	data []byte) (ctypes.BurrowResult, error) {
	if r, err := tmRoutes.tendermintPipe.CallCode(fromAddress, code, data); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func (tmRoutes *TendermintRoutes) DumpStorageResult(address []byte) (ctypes.BurrowResult, error) {
	if r, err := tmRoutes.tendermintPipe.DumpStorage(address); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func (tmRoutes *TendermintRoutes) ListAccountsResult() (ctypes.BurrowResult, error) {
	if r, err := tmRoutes.tendermintPipe.ListAccounts(); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func (tmRoutes *TendermintRoutes) GetNameResult(name string) (ctypes.BurrowResult, error) {
	if r, err := tmRoutes.tendermintPipe.GetName(name); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func (tmRoutes *TendermintRoutes) ListNamesResult() (ctypes.BurrowResult, error) {
	if r, err := tmRoutes.tendermintPipe.ListNames(); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func (tmRoutes *TendermintRoutes) GenPrivAccountResult() (ctypes.BurrowResult, error) {
	//if r, err := tmRoutes.tendermintPipe.GenPrivAccount(); err != nil {
	//	return nil, err
	//} else {
	//	return r, nil
	//}
	return nil, fmt.Errorf("Unimplemented as poor practice to generate private account over unencrypted RPC")
}

func (tmRoutes *TendermintRoutes) SignTxResult(tx txs.Tx,
	privAccounts []*acm.ConcretePrivateAccount) (ctypes.BurrowResult, error) {
	// if r, err := tmRoutes.tendermintPipe.SignTx(tx, privAccounts); err != nil {
	// 	return nil, err
	// } else {
	// 	return r, nil
	// }
	return nil, fmt.Errorf("Unimplemented as poor practice to pass private account over unencrypted RPC")
}

func (tmRoutes *TendermintRoutes) BroadcastTxResult(tx txs.Tx) (ctypes.BurrowResult, error) {
	if r, err := tmRoutes.tendermintPipe.BroadcastTxSync(tx); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func (tmRoutes *TendermintRoutes) BlockchainInfo(minHeight,
	maxHeight int) (ctypes.BurrowResult, error) {
	r, err := tmRoutes.tendermintPipe.BlockchainInfo(minHeight, maxHeight,
		maxBlockLookback)
	if err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func (tmRoutes *TendermintRoutes) ListUnconfirmedTxs() (ctypes.BurrowResult, error) {
	// Get all Txs for now
	r, err := tmRoutes.tendermintPipe.ListUnconfirmedTxs(-1)
	if err != nil {
		return nil, err
	} else {
		return r, nil
	}
}
func (tmRoutes *TendermintRoutes) GetBlock(height int) (ctypes.BurrowResult, error) {
	r, err := tmRoutes.tendermintPipe.GetBlock(height)
	if err != nil {
		return nil, err
	} else {
		return r, nil
	}
}
func (tmRoutes *TendermintRoutes) ListValidators() (ctypes.BurrowResult, error) {
	r, err := tmRoutes.tendermintPipe.ListValidators()
	if err != nil {
		return nil, err
	} else {
		return r, nil
	}
}
func (tmRoutes *TendermintRoutes) DumpConsensusState() (ctypes.BurrowResult, error) {
	return tmRoutes.tendermintPipe.DumpConsensusState()
}
