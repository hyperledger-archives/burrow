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
	"github.com/eris-ltd/eris-db/blockchain"
	core_types "github.com/eris-ltd/eris-db/core/types"
	definitions "github.com/eris-ltd/eris-db/definitions"
	"github.com/eris-ltd/eris-db/event"
	"github.com/eris-ltd/eris-db/rpc"
	"github.com/eris-ltd/eris-db/rpc/v0/shared"
	"github.com/eris-ltd/eris-db/txs"
)

// TODO use the method name definition file.
const (
	SERVICE_NAME = "erisdb"

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
type ErisDbMethods struct {
	codec         rpc.Codec
	pipe          definitions.Pipe
	filterFactory *event.FilterFactory
}

func NewErisDbMethods(codec rpc.Codec,
	pipe definitions.Pipe) *ErisDbMethods {

	return &ErisDbMethods{
		codec:         codec,
		pipe:          pipe,
		filterFactory: blockchain.NewBlockchainFilterFactory(),
	}
}

// Used to handle requests. interface{} param is a wildcard used for example with socket events.
type RequestHandlerFunc func(*rpc.RPCRequest, interface{}) (interface{}, int, error)

// Private. Create a method name -> method handler map.
func (erisDbMethods *ErisDbMethods) getMethods() map[string]RequestHandlerFunc {
	dhMap := make(map[string]RequestHandlerFunc)
	// Accounts
	dhMap[GET_ACCOUNTS] = erisDbMethods.Accounts
	dhMap[GET_ACCOUNT] = erisDbMethods.Account
	dhMap[GET_STORAGE] = erisDbMethods.AccountStorage
	dhMap[GET_STORAGE_AT] = erisDbMethods.AccountStorageAt
	dhMap[GEN_PRIV_ACCOUNT] = erisDbMethods.GenPrivAccount
	dhMap[GEN_PRIV_ACCOUNT_FROM_KEY] = erisDbMethods.GenPrivAccountFromKey
	// Blockchain
	dhMap[GET_BLOCKCHAIN_INFO] = erisDbMethods.BlockchainInfo
	dhMap[GET_GENESIS_HASH] = erisDbMethods.GenesisHash
	dhMap[GET_LATEST_BLOCK_HEIGHT] = erisDbMethods.LatestBlockHeight
	dhMap[GET_LATEST_BLOCK] = erisDbMethods.LatestBlock
	dhMap[GET_BLOCKS] = erisDbMethods.Blocks
	dhMap[GET_BLOCK] = erisDbMethods.Block
	// Consensus
	dhMap[GET_CONSENSUS_STATE] = erisDbMethods.ConsensusState
	dhMap[GET_VALIDATORS] = erisDbMethods.Validators
	// Network
	dhMap[GET_NETWORK_INFO] = erisDbMethods.NetworkInfo
	dhMap[GET_CLIENT_VERSION] = erisDbMethods.ClientVersion
	dhMap[GET_MONIKER] = erisDbMethods.Moniker
	dhMap[GET_CHAIN_ID] = erisDbMethods.ChainId
	dhMap[IS_LISTENING] = erisDbMethods.Listening
	dhMap[GET_LISTENERS] = erisDbMethods.Listeners
	dhMap[GET_PEERS] = erisDbMethods.Peers
	dhMap[GET_PEER] = erisDbMethods.Peer
	// Txs
	dhMap[CALL] = erisDbMethods.Call
	dhMap[CALL_CODE] = erisDbMethods.CallCode
	dhMap[BROADCAST_TX] = erisDbMethods.BroadcastTx
	dhMap[GET_UNCONFIRMED_TXS] = erisDbMethods.UnconfirmedTxs
	dhMap[SIGN_TX] = erisDbMethods.SignTx
	dhMap[TRANSACT] = erisDbMethods.Transact
	dhMap[TRANSACT_AND_HOLD] = erisDbMethods.TransactAndHold
	dhMap[SEND] = erisDbMethods.Send
	dhMap[SEND_AND_HOLD] = erisDbMethods.SendAndHold
	dhMap[TRANSACT_NAMEREG] = erisDbMethods.TransactNameReg
	// Namereg
	dhMap[GET_NAMEREG_ENTRY] = erisDbMethods.NameRegEntry
	dhMap[GET_NAMEREG_ENTRIES] = erisDbMethods.NameRegEntries

	return dhMap
}

// TODO add some sanity checks on address params and such.
// Look into the reflection code in core, see what can be used.

// *************************************** Accounts ***************************************

func (erisDbMethods *ErisDbMethods) GenPrivAccount(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	pac, errC := erisDbMethods.pipe.Accounts().GenPrivAccount()
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return pac, 0, nil
}

func (erisDbMethods *ErisDbMethods) GenPrivAccountFromKey(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {

	param := &PrivKeyParam{}
	err := erisDbMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}

	privKey := param.PrivKey
	pac, errC := erisDbMethods.pipe.Accounts().GenPrivAccountFromKey(privKey)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return pac, 0, nil
}

func (erisDbMethods *ErisDbMethods) Account(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &AddressParam{}
	err := erisDbMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	address := param.Address
	// TODO is address check?
	account, errC := erisDbMethods.pipe.Accounts().Account(address)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return account, 0, nil
}

func (erisDbMethods *ErisDbMethods) Accounts(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &AccountsParam{}
	err := erisDbMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	list, errC := erisDbMethods.pipe.Accounts().Accounts(param.Filters)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return list, 0, nil
}

func (erisDbMethods *ErisDbMethods) AccountStorage(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &AddressParam{}
	err := erisDbMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	address := param.Address
	storage, errC := erisDbMethods.pipe.Accounts().Storage(address)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return storage, 0, nil
}

func (erisDbMethods *ErisDbMethods) AccountStorageAt(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &StorageAtParam{}
	err := erisDbMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	address := param.Address
	key := param.Key
	storageItem, errC := erisDbMethods.pipe.Accounts().StorageAt(address, key)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return storageItem, 0, nil
}

// *************************************** Blockchain ************************************

func (erisDbMethods *ErisDbMethods) BlockchainInfo(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	return shared.BlockchainInfo(erisDbMethods.pipe), 0, nil
}

func (erisDbMethods *ErisDbMethods) ChainId(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	chainId := erisDbMethods.pipe.Blockchain().ChainId()
	return &core_types.ChainId{chainId}, 0, nil
}

func (erisDbMethods *ErisDbMethods) GenesisHash(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	hash := erisDbMethods.pipe.GenesisHash()
	return &core_types.GenesisHash{hash}, 0, nil
}

func (erisDbMethods *ErisDbMethods) LatestBlockHeight(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	height := erisDbMethods.pipe.Blockchain().Height()
	return &core_types.LatestBlockHeight{height}, 0, nil
}

func (erisDbMethods *ErisDbMethods) LatestBlock(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	latestHeight := erisDbMethods.pipe.Blockchain().Height()
	block := erisDbMethods.pipe.Blockchain().Block(latestHeight)
	return block, 0, nil
}

func (erisDbMethods *ErisDbMethods) Blocks(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &BlocksParam{}
	err := erisDbMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	blocks, errC := blockchain.FilterBlocks(erisDbMethods.pipe.Blockchain(), erisDbMethods.filterFactory, param.Filters)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return blocks, 0, nil
}

func (erisDbMethods *ErisDbMethods) Block(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &HeightParam{}
	err := erisDbMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	height := param.Height
	block := erisDbMethods.pipe.Blockchain().Block(height)
	return block, 0, nil
}

// *************************************** Consensus ************************************

func (erisDbMethods *ErisDbMethods) ConsensusState(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	return erisDbMethods.pipe.GetConsensusEngine().ConsensusState(), 0, nil
}

func (erisDbMethods *ErisDbMethods) Validators(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	return erisDbMethods.pipe.GetConsensusEngine().ListValidators(), 0, nil
}

// *************************************** Net ************************************

func (erisDbMethods *ErisDbMethods) NetworkInfo(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	info := shared.NetInfo(erisDbMethods.pipe)
	return info, 0, nil
}

func (erisDbMethods *ErisDbMethods) ClientVersion(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	version := shared.ClientVersion(erisDbMethods.pipe)
	return &core_types.ClientVersion{version}, 0, nil
}

func (erisDbMethods *ErisDbMethods) Moniker(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	moniker := shared.Moniker(erisDbMethods.pipe)
	return &core_types.Moniker{moniker}, 0, nil
}

func (erisDbMethods *ErisDbMethods) Listening(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	listening := shared.Listening(erisDbMethods.pipe)
	return &core_types.Listening{listening}, 0, nil
}

func (erisDbMethods *ErisDbMethods) Listeners(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	listeners := shared.Listeners(erisDbMethods.pipe)
	return &core_types.Listeners{listeners}, 0, nil
}

func (erisDbMethods *ErisDbMethods) Peers(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	peers := erisDbMethods.pipe.GetConsensusEngine().Peers()
	return peers, 0, nil
}

func (erisDbMethods *ErisDbMethods) Peer(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &PeerParam{}
	err := erisDbMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	address := param.Address
	peer := shared.Peer(erisDbMethods.pipe, address)
	return peer, 0, nil
}

// *************************************** Txs ************************************

func (erisDbMethods *ErisDbMethods) Call(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &CallParam{}
	err := erisDbMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	from := param.From
	to := param.Address
	data := param.Data
	call, errC := erisDbMethods.pipe.Transactor().Call(from, to, data)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return call, 0, nil
}

func (erisDbMethods *ErisDbMethods) CallCode(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &CallCodeParam{}
	err := erisDbMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	from := param.From
	code := param.Code
	data := param.Data
	call, errC := erisDbMethods.pipe.Transactor().CallCode(from, code, data)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return call, 0, nil
}

func (erisDbMethods *ErisDbMethods) BroadcastTx(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	tx := new(txs.Tx)
	err := erisDbMethods.codec.DecodeBytes(tx, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	receipt, errC := erisDbMethods.pipe.Transactor().BroadcastTx(*tx)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return receipt, 0, nil
}

func (erisDbMethods *ErisDbMethods) Transact(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &TransactParam{}
	err := erisDbMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	receipt, errC := erisDbMethods.pipe.Transactor().Transact(param.PrivKey, param.Address, param.Data, param.GasLimit, param.Fee)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return receipt, 0, nil
}

func (erisDbMethods *ErisDbMethods) TransactAndHold(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &TransactParam{}
	err := erisDbMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	ce, errC := erisDbMethods.pipe.Transactor().TransactAndHold(param.PrivKey, param.Address, param.Data, param.GasLimit, param.Fee)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return ce, 0, nil
}

func (this *ErisDbMethods) Send(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
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

func (this *ErisDbMethods) SendAndHold(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
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

func (erisDbMethods *ErisDbMethods) TransactNameReg(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &TransactNameRegParam{}
	err := erisDbMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	receipt, errC := erisDbMethods.pipe.Transactor().TransactNameReg(param.PrivKey, param.Name, param.Data, param.Amount, param.Fee)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return receipt, 0, nil
}

func (erisDbMethods *ErisDbMethods) UnconfirmedTxs(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	trans, errC := erisDbMethods.pipe.GetConsensusEngine().ListUnconfirmedTxs(-1)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return txs.UnconfirmedTxs{trans}, 0, nil
}

func (erisDbMethods *ErisDbMethods) SignTx(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &SignTxParam{}
	err := erisDbMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	tx := param.Tx
	pAccs := param.PrivAccounts
	txRet, errC := erisDbMethods.pipe.Transactor().SignTx(tx, pAccs)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return txRet, 0, nil
}

// *************************************** Name Registry ***************************************

func (erisDbMethods *ErisDbMethods) NameRegEntry(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &NameRegEntryParam{}
	err := erisDbMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	name := param.Name
	// TODO is address check?
	entry, errC := erisDbMethods.pipe.NameReg().Entry(name)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return entry, 0, nil
}

func (erisDbMethods *ErisDbMethods) NameRegEntries(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &FilterListParam{}
	err := erisDbMethods.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	list, errC := erisDbMethods.pipe.NameReg().Entries(param.Filters)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return list, 0, nil
}
