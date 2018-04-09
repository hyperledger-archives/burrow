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

package v0

import (
	"fmt"
	"os"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/rpc/filters"
	"github.com/hyperledger/burrow/txs"
)

// TODO use the method name definition file.
const (
	SERVICE_NAME = "burrow"

	GET_ACCOUNTS              = SERVICE_NAME + ".getAccounts" // Accounts
	GET_ACCOUNT               = SERVICE_NAME + ".getAccount"
	GET_STORAGE               = SERVICE_NAME + ".getStorage"
	GET_STORAGE_AT            = SERVICE_NAME + ".getStorageAt"
	GEN_PRIV_ACCOUNT          = SERVICE_NAME + ".genPrivAccount"
	GEN_PRIV_ACCOUNT_FROM_KEY = SERVICE_NAME + ".genPrivAccountFromKey"
	GET_BLOCKCHAIN_INFO       = SERVICE_NAME + ".getBlockchainInfo" // Blockchain
	GET_LATEST_BLOCK          = SERVICE_NAME + ".getLatestBlock"
	GET_BLOCKS                = SERVICE_NAME + ".getBlocks"
	GET_BLOCK                 = SERVICE_NAME + ".getBlock"
	GET_CONSENSUS_STATE       = SERVICE_NAME + ".getConsensusState" // Consensus
	GET_VALIDATORS            = SERVICE_NAME + ".getValidators"
	GET_NETWORK_INFO          = SERVICE_NAME + ".getNetworkInfo" // Net
	GET_CHAIN_ID              = SERVICE_NAME + ".getChainId"
	GET_PEERS                 = SERVICE_NAME + ".getPeers"
	CALL                      = SERVICE_NAME + ".call" // Tx
	CALL_CODE                 = SERVICE_NAME + ".callCode"
	BROADCAST_TX              = SERVICE_NAME + ".broadcastTx"
	GET_UNCONFIRMED_TXS       = SERVICE_NAME + ".getUnconfirmedTxs"
	SIGN_TX                   = SERVICE_NAME + ".signTx"
	TRANSACT                  = SERVICE_NAME + ".transact"
	TRANSACT_AND_HOLD         = SERVICE_NAME + ".transactAndHold"
	SEND                      = SERVICE_NAME + ".send"
	SEND_AND_HOLD             = SERVICE_NAME + ".sendAndHold"
	TRANSACT_NAMEREG          = SERVICE_NAME + ".transactNameReg"
	EVENT_SUBSCRIBE           = SERVICE_NAME + ".eventSubscribe" // Events
	EVENT_UNSUBSCRIBE         = SERVICE_NAME + ".eventUnsubscribe"
	EVENT_POLL                = SERVICE_NAME + ".eventPoll"
	GET_NAMEREG_ENTRY         = SERVICE_NAME + ".getNameRegEntry" // Namereg
	GET_NAMEREG_ENTRIES       = SERVICE_NAME + ".getNameRegEntries"
)

// Used to handle requests. interface{} param is a wildcard used for example with socket events.
type RequestHandlerFunc func(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error)

// Private. Create a method name -> method handler map.
func GetMethods(codec rpc.Codec, service *rpc.Service, logger *logging.Logger) map[string]RequestHandlerFunc {
	accountFilterFactory := filters.NewAccountFilterFactory()
	nameRegFilterFactory := filters.NewNameRegFilterFactory()
	return map[string]RequestHandlerFunc{
		// Accounts
		GET_ACCOUNTS: func(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
			param := &FilterListParam{}
			if len(request.Params) > 0 {
				err := codec.DecodeBytes(param, request.Params)
				if err != nil {
					return nil, rpc.INVALID_PARAMS, err
				}
			}
			filter, err := accountFilterFactory.NewFilter(param.Filters)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			list, err := service.ListAccounts(func(account acm.Account) bool {
				return filter.Match(account)

			})
			if err != nil {
				return nil, rpc.INTERNAL_ERROR, err
			}
			return list, 0, nil
		},
		GET_ACCOUNT: func(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
			param := &AddressParam{}
			err := codec.DecodeBytes(param, request.Params)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			address, err := acm.AddressFromBytes(param.Address)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			account, err := service.GetAccount(address)
			if err != nil {
				return nil, rpc.INTERNAL_ERROR, err
			}
			return account, 0, nil
		},
		GET_STORAGE: func(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
			param := &AddressParam{}
			err := codec.DecodeBytes(param, request.Params)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			address, err := acm.AddressFromBytes(param.Address)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			storage, err := service.DumpStorage(address)
			if err != nil {
				return nil, rpc.INTERNAL_ERROR, err
			}
			return storage, 0, nil
		},
		GET_STORAGE_AT: func(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
			param := &StorageAtParam{}
			err := codec.DecodeBytes(param, request.Params)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			address, err := acm.AddressFromBytes(param.Address)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			storageItem, err := service.GetStorage(address, param.Key)
			if err != nil {
				return nil, rpc.INTERNAL_ERROR, err
			}
			return storageItem, 0, nil
		},
		GEN_PRIV_ACCOUNT: func(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
			pa, err := acm.GeneratePrivateAccount()
			if err != nil {
				return nil, rpc.INTERNAL_ERROR, err
			}
			return acm.AsConcretePrivateAccount(pa), 0, nil
		},
		GEN_PRIV_ACCOUNT_FROM_KEY: func(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
			param := &PrivKeyParam{}
			err := codec.DecodeBytes(param, request.Params)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			pa, err := acm.GeneratePrivateAccountFromPrivateKeyBytes(param.PrivKey)
			if err != nil {
				return nil, rpc.INTERNAL_ERROR, err
			}
			return acm.AsConcretePrivateAccount(pa), 0, nil
		},
		// Txs
		CALL: func(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
			param := &CallParam{}
			err := codec.DecodeBytes(param, request.Params)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			from, err := acm.AddressFromBytes(param.From)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			address, err := acm.AddressFromBytes(param.Address)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			call, err := service.Transactor().Call(service.MempoolAccounts(), from, address, param.Data)
			if err != nil {
				return nil, rpc.INTERNAL_ERROR, err
			}
			return call, 0, nil
		},
		CALL_CODE: func(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
			param := &CallCodeParam{}
			err := codec.DecodeBytes(param, request.Params)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			from, err := acm.AddressFromBytes(param.From)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			call, err := service.Transactor().CallCode(service.MempoolAccounts(), from, param.Code, param.Data)
			if err != nil {
				return nil, rpc.INTERNAL_ERROR, err
			}
			return call, 0, nil
		},
		BROADCAST_TX: func(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
			// Accept all transaction types as parameter for broadcast.
			param := new(txs.Tx)
			err := codec.DecodeBytesPtr(param, request.Params)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			receipt, err := service.Transactor().BroadcastTx(*param)
			if err != nil {
				return nil, rpc.INTERNAL_ERROR, err
			}
			return receipt, 0, nil
		},
		SIGN_TX: func(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
			param := &SignTxParam{}
			err := codec.DecodeBytes(param, request.Params)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			txRet, err := service.Transactor().SignTx(param.Tx, acm.SigningAccounts(param.PrivAccounts))
			if err != nil {
				return nil, rpc.INTERNAL_ERROR, err
			}
			return txRet, 0, nil
		},
		TRANSACT: func(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
			fmt.Fprintf(os.Stderr, "Got transact\n")
			param := &TransactParam{}
			err := codec.DecodeBytes(param, request.Params)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			address, err := acm.MaybeAddressFromBytes(param.Address)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			// Use mempool state so that transact can generate a run of sequence numbers when formulating transactions
			inputAccount, err := signingAccount(service.MempoolAccounts(), param.PrivKey, param.InputAddress)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			receipt, err := service.Transactor().Transact(inputAccount, address, param.Data, param.GasLimit, param.Fee)
			if err != nil {
				return nil, rpc.INTERNAL_ERROR, err
			}
			return receipt, 0, nil
		},
		TRANSACT_AND_HOLD: func(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
			param := &TransactParam{}
			err := codec.DecodeBytes(param, request.Params)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			address, err := acm.MaybeAddressFromBytes(param.Address)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			inputAccount, err := signingAccount(service.MempoolAccounts(), param.PrivKey, param.InputAddress)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			ce, err := service.Transactor().TransactAndHold(inputAccount, address, param.Data, param.GasLimit, param.Fee)
			if err != nil {
				return nil, rpc.INTERNAL_ERROR, err
			}
			return ce, 0, nil
		},
		SEND: func(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
			param := &SendParam{}
			err := codec.DecodeBytes(param, request.Params)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			toAddress, err := acm.AddressFromBytes(param.ToAddress)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			// Run Send against mempool state
			inputAccount, err := signingAccount(service.MempoolAccounts(), param.PrivKey, param.InputAddress)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			receipt, err := service.Transactor().Send(inputAccount, toAddress, param.Amount)
			if err != nil {
				return nil, rpc.INTERNAL_ERROR, err
			}
			return receipt, 0, nil
		},
		SEND_AND_HOLD: func(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
			param := &SendParam{}
			err := codec.DecodeBytes(param, request.Params)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			toAddress, err := acm.AddressFromBytes(param.ToAddress)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			// Run Send against mempool state
			inputAccount, err := signingAccount(service.MempoolAccounts(), param.PrivKey, param.InputAddress)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			rec, err := service.Transactor().SendAndHold(inputAccount, toAddress, param.Amount)
			if err != nil {
				return nil, rpc.INTERNAL_ERROR, err
			}
			return rec, 0, nil
		},
		TRANSACT_NAMEREG: func(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
			param := &TransactNameRegParam{}
			err := codec.DecodeBytes(param, request.Params)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			inputAccount, err := signingAccount(service.MempoolAccounts(), param.PrivKey, param.InputAddress)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			receipt, err := service.Transactor().TransactNameReg(inputAccount, param.Name, param.Data, param.Amount, param.Fee)
			if err != nil {
				return nil, rpc.INTERNAL_ERROR, err
			}
			return receipt, 0, nil
		},

		// Namereg
		GET_NAMEREG_ENTRY: func(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
			param := &NameRegEntryParam{}
			err := codec.DecodeBytes(param, request.Params)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			name := param.Name
			// TODO is address check?
			resultGetName, err := service.GetName(name)
			if err != nil {
				return nil, rpc.INTERNAL_ERROR, err
			}
			return resultGetName.Entry, 0, nil
		},
		GET_NAMEREG_ENTRIES: func(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
			param := &FilterListParam{}
			err := codec.DecodeBytes(param, request.Params)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			filter, err := nameRegFilterFactory.NewFilter(param.Filters)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			list, err := service.ListNames(func(entry *execution.NameRegEntry) bool {
				return filter.Match(entry)
			})
			if err != nil {
				return nil, rpc.INTERNAL_ERROR, err
			}
			return list, 0, nil
		},
		// Blockchain
		GET_BLOCKCHAIN_INFO: func(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
			resultStatus, err := service.Status()
			if err != nil {
				return nil, rpc.INTERNAL_ERROR, err
			}
			return resultStatus, 0, nil
		},
		GET_LATEST_BLOCK: func(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
			stat, err := service.Status()
			if err != nil {
				return nil, rpc.INTERNAL_ERROR, err
			}
			resultGetBlock, err := service.GetBlock(stat.LatestBlockHeight)
			if err != nil {
				return nil, rpc.INTERNAL_ERROR, err
			}
			return resultGetBlock, 0, nil
		},
		GET_BLOCKS: func(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
			param := &BlocksParam{}
			err := codec.DecodeBytes(param, request.Params)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			blocks, err := service.ListBlocks(param.MinHeight, param.MaxHeight)
			if err != nil {
				return nil, rpc.INTERNAL_ERROR, err
			}
			return blocks, 0, nil
		},
		GET_BLOCK: func(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
			param := &HeightParam{}
			err := codec.DecodeBytes(param, request.Params)
			if err != nil {
				return nil, rpc.INVALID_PARAMS, err
			}
			block, err := service.GetBlock(param.Height)
			if err != nil {
				return nil, rpc.INTERNAL_ERROR, err
			}
			return block, 0, nil
		},
		GET_UNCONFIRMED_TXS: func(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
			trans, err := service.ListUnconfirmedTxs(-1)
			if err != nil {
				return nil, rpc.INTERNAL_ERROR, err
			}
			return trans, 0, nil
		},
		// Consensus
		GET_CONSENSUS_STATE: func(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
			resultDumpConsensusState, err := service.DumpConsensusState()
			if err != nil {
				return nil, rpc.INTERNAL_ERROR, err
			}
			return resultDumpConsensusState, 0, nil
		},
		GET_VALIDATORS: func(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
			resultListValidators, err := service.ListValidators()
			if err != nil {
				return nil, rpc.INTERNAL_ERROR, err
			}
			return resultListValidators, 0, nil
		},
		// Network
		GET_NETWORK_INFO: func(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
			resultNetInfo, err := service.NetInfo()
			if err != nil {
				return nil, rpc.INTERNAL_ERROR, err
			}
			return resultNetInfo, 0, nil
		},
		GET_CHAIN_ID: func(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
			resultChainID, err := service.ChainId()
			if err != nil {
				return nil, rpc.INTERNAL_ERROR, err
			}
			return resultChainID, 0, nil
		},
		GET_PEERS: func(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
			resultPeers, err := service.Peers()
			if err != nil {
				return nil, rpc.INTERNAL_ERROR, err
			}
			return resultPeers, 0, nil
		},
	}
}

// Gets signing account from onr of private key or address - failing if both are provided
func signingAccount(accounts *execution.Accounts, privKey, addressBytes []byte) (*execution.SequentialSigningAccount, error) {
	if len(addressBytes) > 0 {
		if len(privKey) > 0 {
			return nil, fmt.Errorf("privKey and address provided but only one or the other should be given")
		}
		address, err := acm.AddressFromBytes(addressBytes)
		if err != nil {
			return nil, err
		}
		return accounts.SequentialSigningAccount(address), nil
	}

	return accounts.SequentialSigningAccountFromPrivateKey(privKey)
}
