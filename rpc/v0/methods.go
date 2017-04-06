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

package rpc_v0

import (
	"github.com/monax/burrow/blockchain"
	core_types "github.com/monax/burrow/core/types"
	definitions "github.com/monax/burrow/definitions"
	"github.com/monax/burrow/event"
	"github.com/monax/burrow/rpc"
	"github.com/monax/burrow/rpc/v0/shared"
	"github.com/monax/burrow/txs"
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
	GET_GENESIS_HASH          = SERVICE_NAME + ".getGenesisHash"
	GET_LATEST_BLOCK_HEIGHT   = SERVICE_NAME + ".getLatestBlockHeight"
	GET_LATEST_BLOCK          = SERVICE_NAME + ".getLatestBlock"
	GET_BLOCKS                = SERVICE_NAME + ".getBlocks"
	GET_BLOCK                 = SERVICE_NAME + ".getBlock"
	GET_CONSENSUS_STATE       = SERVICE_NAME + ".getConsensusState" // Consensus
	GET_VALIDATORS            = SERVICE_NAME + ".getValidators"
	GET_NETWORK_INFO          = SERVICE_NAME + ".getNetworkInfo" // Net
	GET_CLIENT_VERSION        = SERVICE_NAME + ".getClientVersion"
	GET_MONIKER               = SERVICE_NAME + ".getMoniker"
	GET_CHAIN_ID              = SERVICE_NAME + ".getChainId"
	IS_LISTENING              = SERVICE_NAME + ".isListening"
	GET_LISTENERS             = SERVICE_NAME + ".getListeners"
	GET_PEERS                 = SERVICE_NAME + ".getPeers"
	GET_PEER                  = SERVICE_NAME + ".getPeer"
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

// The rpc method handlers.
type BurrowMethods struct {
	codec         rpc.Codec
	pipe          definitions.Pipe
	filterFactory *event.FilterFactory
}

func NewBurrowMethods(codec rpc.Codec,
	pipe definitions.Pipe) *BurrowMethods {

	return &BurrowMethods{
		codec:         codec,
		pipe:          pipe,
		filterFactory: blockchain.NewBlockchainFilterFactory(),
	}
}

// Used to handle requests. interface{} param is a wildcard used for example with socket events.
type RequestHandlerFunc func(*rpc.RPCRequest, interface{}) (interface{}, int, error)

// Private. Create a method name -> method handler map.
func (burrowMethods *BurrowMethods) getMethods() map[string]RequestHandlerFunc {
	dhMap := make(map[string]RequestHandlerFunc)
	// Accounts
	dhMap[GET_ACCOUNTS] = burrowMethods.Accounts
	dhMap[GET_ACCOUNT] = burrowMethods.Account
	dhMap[GET_STORAGE] = burrowMethods.AccountStorage
	dhMap[GET_STORAGE_AT] = burrowMethods.AccountStorageAt
	dhMap[GEN_PRIV_ACCOUNT] = burrowMethods.GenPrivAccount
	dhMap[GEN_PRIV_ACCOUNT_FROM_KEY] = burrowMethods.GenPrivAccountFromKey
	// Blockchain
	dhMap[GET_BLOCKCHAIN_INFO] = burrowMethods.BlockchainInfo
	dhMap[GET_GENESIS_HASH] = burrowMethods.GenesisHash
	dhMap[GET_LATEST_BLOCK_HEIGHT] = burrowMethods.LatestBlockHeight
	dhMap[GET_LATEST_BLOCK] = burrowMethods.LatestBlock
	dhMap[GET_BLOCKS] = burrowMethods.Blocks
	dhMap[GET_BLOCK] = burrowMethods.Block
	// Consensus
	dhMap[GET_CONSENSUS_STATE] = burrowMethods.ConsensusState
	dhMap[GET_VALIDATORS] = burrowMethods.Validators
	// Network
	dhMap[GET_NETWORK_INFO] = burrowMethods.NetworkInfo
	dhMap[GET_CLIENT_VERSION] = burrowMethods.ClientVersion
	dhMap[GET_MONIKER] = burrowMethods.Moniker
	dhMap[GET_CHAIN_ID] = burrowMethods.ChainId
	dhMap[IS_LISTENING] = burrowMethods.Listening
	dhMap[GET_LISTENERS] = burrowMethods.Listeners
	dhMap[GET_PEERS] = burrowMethods.Peers
	dhMap[GET_PEER] = burrowMethods.Peer
	// Txs
	dhMap[CALL] = burrowMethods.Call
	dhMap[CALL_CODE] = burrowMethods.CallCode
	dhMap[BROADCAST_TX] = burrowMethods.BroadcastTx
	dhMap[GET_UNCONFIRMED_TXS] = burrowMethods.UnconfirmedTxs
	dhMap[SIGN_TX] = burrowMethods.SignTx
	dhMap[TRANSACT] = burrowMethods.Transact
	dhMap[TRANSACT_AND_HOLD] = burrowMethods.TransactAndHold
	dhMap[SEND] = burrowMethods.Send
	dhMap[SEND_AND_HOLD] = burrowMethods.SendAndHold
	dhMap[TRANSACT_NAMEREG] = burrowMethods.TransactNameReg
	// Namereg
	dhMap[GET_NAMEREG_ENTRY] = burrowMethods.NameRegEntry
	dhMap[GET_NAMEREG_ENTRIES] = burrowMethods.NameRegEntries

	return dhMap
}

// TODO add some sanity checks on address params and such.
// Look into the reflection code in core, see what can be used.

// *************************************** Accounts ***************************************

func (burrowMethods *BurrowMethods) GenPrivAccount(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	pac, errC := burrowMethods.pipe.Accounts().GenPrivAccount()
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return pac, 0, nil
}

func (burrowMethods *BurrowMethods) GenPrivAccountFromKey(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {

	param := &PrivKeyParam{}
	err := burrowMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}

	privKey := param.PrivKey
	pac, errC := burrowMethods.pipe.Accounts().GenPrivAccountFromKey(privKey)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return pac, 0, nil
}

func (burrowMethods *BurrowMethods) Account(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &AddressParam{}
	err := burrowMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	address := param.Address
	// TODO is address check?
	account, errC := burrowMethods.pipe.Accounts().Account(address)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return account, 0, nil
}

func (burrowMethods *BurrowMethods) Accounts(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &AccountsParam{}
	err := burrowMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	list, errC := burrowMethods.pipe.Accounts().Accounts(param.Filters)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return list, 0, nil
}

func (burrowMethods *BurrowMethods) AccountStorage(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &AddressParam{}
	err := burrowMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	address := param.Address
	storage, errC := burrowMethods.pipe.Accounts().Storage(address)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return storage, 0, nil
}

func (burrowMethods *BurrowMethods) AccountStorageAt(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &StorageAtParam{}
	err := burrowMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	address := param.Address
	key := param.Key
	storageItem, errC := burrowMethods.pipe.Accounts().StorageAt(address, key)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return storageItem, 0, nil
}

// *************************************** Blockchain ************************************

func (burrowMethods *BurrowMethods) BlockchainInfo(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	return shared.BlockchainInfo(burrowMethods.pipe), 0, nil
}

func (burrowMethods *BurrowMethods) ChainId(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	chainId := burrowMethods.pipe.Blockchain().ChainId()
	return &core_types.ChainId{chainId}, 0, nil
}

func (burrowMethods *BurrowMethods) GenesisHash(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	hash := burrowMethods.pipe.GenesisHash()
	return &core_types.GenesisHash{hash}, 0, nil
}

func (burrowMethods *BurrowMethods) LatestBlockHeight(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	height := burrowMethods.pipe.Blockchain().Height()
	return &core_types.LatestBlockHeight{height}, 0, nil
}

func (burrowMethods *BurrowMethods) LatestBlock(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	latestHeight := burrowMethods.pipe.Blockchain().Height()
	block := burrowMethods.pipe.Blockchain().Block(latestHeight)
	return block, 0, nil
}

func (burrowMethods *BurrowMethods) Blocks(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &BlocksParam{}
	err := burrowMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	blocks, errC := blockchain.FilterBlocks(burrowMethods.pipe.Blockchain(), burrowMethods.filterFactory, param.Filters)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return blocks, 0, nil
}

func (burrowMethods *BurrowMethods) Block(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &HeightParam{}
	err := burrowMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	height := param.Height
	block := burrowMethods.pipe.Blockchain().Block(height)
	return block, 0, nil
}

// *************************************** Consensus ************************************

func (burrowMethods *BurrowMethods) ConsensusState(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	return burrowMethods.pipe.GetConsensusEngine().ConsensusState(), 0, nil
}

func (burrowMethods *BurrowMethods) Validators(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	return burrowMethods.pipe.GetConsensusEngine().ListValidators(), 0, nil
}

// *************************************** Net ************************************

func (burrowMethods *BurrowMethods) NetworkInfo(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	info := shared.NetInfo(burrowMethods.pipe)
	return info, 0, nil
}

func (burrowMethods *BurrowMethods) ClientVersion(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	version := shared.ClientVersion(burrowMethods.pipe)
	return &core_types.ClientVersion{version}, 0, nil
}

func (burrowMethods *BurrowMethods) Moniker(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	moniker := shared.Moniker(burrowMethods.pipe)
	return &core_types.Moniker{moniker}, 0, nil
}

func (burrowMethods *BurrowMethods) Listening(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	listening := shared.Listening(burrowMethods.pipe)
	return &core_types.Listening{listening}, 0, nil
}

func (burrowMethods *BurrowMethods) Listeners(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	listeners := shared.Listeners(burrowMethods.pipe)
	return &core_types.Listeners{listeners}, 0, nil
}

func (burrowMethods *BurrowMethods) Peers(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	peers := burrowMethods.pipe.GetConsensusEngine().Peers()
	return peers, 0, nil
}

func (burrowMethods *BurrowMethods) Peer(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &PeerParam{}
	err := burrowMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	address := param.Address
	peer := shared.Peer(burrowMethods.pipe, address)
	return peer, 0, nil
}

// *************************************** Txs ************************************

func (burrowMethods *BurrowMethods) Call(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &CallParam{}
	err := burrowMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	from := param.From
	to := param.Address
	data := param.Data
	call, errC := burrowMethods.pipe.Transactor().Call(from, to, data)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return call, 0, nil
}

func (burrowMethods *BurrowMethods) CallCode(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &CallCodeParam{}
	err := burrowMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	from := param.From
	code := param.Code
	data := param.Data
	call, errC := burrowMethods.pipe.Transactor().CallCode(from, code, data)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return call, 0, nil
}

func (burrowMethods *BurrowMethods) BroadcastTx(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	// Accept all transaction types as parameter for broadcast.
	param := new(txs.Tx)
	err := burrowMethods.codec.DecodeBytesPtr(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	receipt, errC := burrowMethods.pipe.Transactor().BroadcastTx(*param)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return receipt, 0, nil
}

func (burrowMethods *BurrowMethods) Transact(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &TransactParam{}
	err := burrowMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	receipt, errC := burrowMethods.pipe.Transactor().Transact(param.PrivKey, param.Address, param.Data, param.GasLimit, param.Fee)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return receipt, 0, nil
}

func (burrowMethods *BurrowMethods) TransactAndHold(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &TransactParam{}
	err := burrowMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	ce, errC := burrowMethods.pipe.Transactor().TransactAndHold(param.PrivKey, param.Address, param.Data, param.GasLimit, param.Fee)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return ce, 0, nil
}

func (this *BurrowMethods) Send(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &SendParam{}
	err := this.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	receipt, errC := this.pipe.Transactor().Send(param.PrivKey, param.ToAddress, param.Amount)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return receipt, 0, nil
}

func (this *BurrowMethods) SendAndHold(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &SendParam{}
	err := this.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	rec, errC := this.pipe.Transactor().SendAndHold(param.PrivKey, param.ToAddress, param.Amount)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return rec, 0, nil
}

func (burrowMethods *BurrowMethods) TransactNameReg(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &TransactNameRegParam{}
	err := burrowMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	receipt, errC := burrowMethods.pipe.Transactor().TransactNameReg(param.PrivKey, param.Name, param.Data, param.Amount, param.Fee)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return receipt, 0, nil
}

func (burrowMethods *BurrowMethods) UnconfirmedTxs(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	trans, errC := burrowMethods.pipe.GetConsensusEngine().ListUnconfirmedTxs(-1)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return txs.UnconfirmedTxs{trans}, 0, nil
}

func (burrowMethods *BurrowMethods) SignTx(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &SignTxParam{}
	err := burrowMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	tx := param.Tx
	pAccs := param.PrivAccounts
	txRet, errC := burrowMethods.pipe.Transactor().SignTx(tx, pAccs)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return txRet, 0, nil
}

// *************************************** Name Registry ***************************************

func (burrowMethods *BurrowMethods) NameRegEntry(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &NameRegEntryParam{}
	err := burrowMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	name := param.Name
	// TODO is address check?
	entry, errC := burrowMethods.pipe.NameReg().Entry(name)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return entry, 0, nil
}

func (burrowMethods *BurrowMethods) NameRegEntries(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &FilterListParam{}
	err := burrowMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	list, errC := burrowMethods.pipe.NameReg().Entries(param.Filters)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return list, 0, nil
}
