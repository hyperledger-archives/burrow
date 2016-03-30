package erisdb

import (
	"crypto/rand"
	"encoding/hex"
	"strings"

	ep "github.com/eris-ltd/eris-db/erisdb/pipe"
	rpc "github.com/eris-ltd/eris-db/rpc"

	"github.com/eris-ltd/eris-db/tendermint/tendermint/types"
	"github.com/eris-ltd/eris-db/tendermint/tendermint/wire"
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
	codec rpc.Codec
	pipe  ep.Pipe
}

// Used to handle requests. interface{} param is a wildcard used for example with socket events.
type RequestHandlerFunc func(*rpc.RPCRequest, interface{}) (interface{}, int, error)

// Private. Create a method name -> method handler map.
func (this *ErisDbMethods) getMethods() map[string]RequestHandlerFunc {
	dhMap := make(map[string]RequestHandlerFunc)
	// Accounts
	dhMap[GET_ACCOUNTS] = this.Accounts
	dhMap[GET_ACCOUNT] = this.Account
	dhMap[GET_STORAGE] = this.AccountStorage
	dhMap[GET_STORAGE_AT] = this.AccountStorageAt
	dhMap[GEN_PRIV_ACCOUNT] = this.GenPrivAccount
	dhMap[GEN_PRIV_ACCOUNT_FROM_KEY] = this.GenPrivAccountFromKey
	// Blockchain
	dhMap[GET_BLOCKCHAIN_INFO] = this.BlockchainInfo
	dhMap[GET_GENESIS_HASH] = this.GenesisHash
	dhMap[GET_LATEST_BLOCK_HEIGHT] = this.LatestBlockHeight
	dhMap[GET_LATEST_BLOCK] = this.LatestBlock
	dhMap[GET_BLOCKS] = this.Blocks
	dhMap[GET_BLOCK] = this.Block
	// Consensus
	dhMap[GET_CONSENSUS_STATE] = this.ConsensusState
	dhMap[GET_VALIDATORS] = this.Validators
	// Network
	dhMap[GET_NETWORK_INFO] = this.NetworkInfo
	dhMap[GET_CLIENT_VERSION] = this.ClientVersion
	dhMap[GET_MONIKER] = this.Moniker
	dhMap[GET_CHAIN_ID] = this.ChainId
	dhMap[IS_LISTENING] = this.Listening
	dhMap[GET_LISTENERS] = this.Listeners
	dhMap[GET_PEERS] = this.Peers
	dhMap[GET_PEER] = this.Peer
	// Txs
	dhMap[CALL] = this.Call
	dhMap[CALL_CODE] = this.CallCode
	dhMap[BROADCAST_TX] = this.BroadcastTx
	dhMap[GET_UNCONFIRMED_TXS] = this.UnconfirmedTxs
	dhMap[SIGN_TX] = this.SignTx
	dhMap[TRANSACT] = this.Transact
	dhMap[TRANSACT_AND_HOLD] = this.TransactAndHold
	dhMap[SEND] = this.Send
	dhMap[SEND_AND_HOLD] = this.SendAndHold
	dhMap[TRANSACT_NAMEREG] = this.TransactNameReg
	// Namereg
	dhMap[GET_NAMEREG_ENTRY] = this.NameRegEntry
	dhMap[GET_NAMEREG_ENTRIES] = this.NameRegEntries

	return dhMap
}

// TODO add some sanity checks on address params and such.
// Look into the reflection code in core, see what can be used.

// *************************************** Accounts ***************************************

func (this *ErisDbMethods) GenPrivAccount(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	pac, errC := this.pipe.Accounts().GenPrivAccount()
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return pac, 0, nil
}

func (this *ErisDbMethods) GenPrivAccountFromKey(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {

	param := &PrivKeyParam{}
	err := this.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}

	privKey := param.PrivKey
	pac, errC := this.pipe.Accounts().GenPrivAccountFromKey(privKey)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return pac, 0, nil
}

func (this *ErisDbMethods) Account(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &AddressParam{}
	err := this.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	address := param.Address
	// TODO is address check?
	account, errC := this.pipe.Accounts().Account(address)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return account, 0, nil
}

func (this *ErisDbMethods) Accounts(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &AccountsParam{}
	err := this.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	list, errC := this.pipe.Accounts().Accounts(param.Filters)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return list, 0, nil
}

func (this *ErisDbMethods) AccountStorage(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &AddressParam{}
	err := this.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	address := param.Address
	storage, errC := this.pipe.Accounts().Storage(address)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return storage, 0, nil
}

func (this *ErisDbMethods) AccountStorageAt(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &StorageAtParam{}
	err := this.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	address := param.Address
	key := param.Key
	storageItem, errC := this.pipe.Accounts().StorageAt(address, key)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return storageItem, 0, nil
}

// *************************************** Blockchain ************************************

func (this *ErisDbMethods) BlockchainInfo(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	status, errC := this.pipe.Blockchain().Info()
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return status, 0, nil
}

func (this *ErisDbMethods) ChainId(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	chainId, errC := this.pipe.Blockchain().ChainId()
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return &ep.ChainId{chainId}, 0, nil
}

func (this *ErisDbMethods) GenesisHash(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	hash, errC := this.pipe.Blockchain().GenesisHash()
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return &ep.GenesisHash{hash}, 0, nil
}

func (this *ErisDbMethods) LatestBlockHeight(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	height, errC := this.pipe.Blockchain().LatestBlockHeight()
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return &ep.LatestBlockHeight{height}, 0, nil
}

func (this *ErisDbMethods) LatestBlock(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {

	block, errC := this.pipe.Blockchain().LatestBlock()
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return block, 0, nil
}

func (this *ErisDbMethods) Blocks(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &BlocksParam{}
	err := this.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	blocks, errC := this.pipe.Blockchain().Blocks(param.Filters)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return blocks, 0, nil
}

func (this *ErisDbMethods) Block(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &HeightParam{}
	err := this.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	height := param.Height
	block, errC := this.pipe.Blockchain().Block(height)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return block, 0, nil
}

// *************************************** Consensus ************************************

func (this *ErisDbMethods) ConsensusState(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	state, errC := this.pipe.Consensus().State()
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return state, 0, nil
}

func (this *ErisDbMethods) Validators(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	validators, errC := this.pipe.Consensus().Validators()
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return validators, 0, nil
}

// *************************************** Net ************************************

func (this *ErisDbMethods) NetworkInfo(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	info, errC := this.pipe.Net().Info()
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return info, 0, nil
}

func (this *ErisDbMethods) ClientVersion(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	version, errC := this.pipe.Net().ClientVersion()
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return &ep.ClientVersion{version}, 0, nil
}

func (this *ErisDbMethods) Moniker(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	moniker, errC := this.pipe.Net().Moniker()
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return &ep.Moniker{moniker}, 0, nil
}

func (this *ErisDbMethods) Listening(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	listening, errC := this.pipe.Net().Listening()
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return &ep.Listening{listening}, 0, nil
}

func (this *ErisDbMethods) Listeners(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	listeners, errC := this.pipe.Net().Listeners()
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return &ep.Listeners{listeners}, 0, nil
}

func (this *ErisDbMethods) Peers(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	peers, errC := this.pipe.Net().Peers()
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return peers, 0, nil
}

func (this *ErisDbMethods) Peer(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &PeerParam{}
	err := this.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	address := param.Address
	peer, errC := this.pipe.Net().Peer(address)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return peer, 0, nil
}

// *************************************** Txs ************************************

func (this *ErisDbMethods) Call(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &CallParam{}
	err := this.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	from := param.From
	to := param.Address
	data := param.Data
	call, errC := this.pipe.Transactor().Call(from, to, data)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return call, 0, nil
}

func (this *ErisDbMethods) CallCode(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &CallCodeParam{}
	err := this.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	from := param.From
	code := param.Code
	data := param.Data
	call, errC := this.pipe.Transactor().CallCode(from, code, data)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return call, 0, nil
}

func (this *ErisDbMethods) BroadcastTx(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	var err error
	// Special because Tx is an interface
	param := new(types.Tx)
	wire.ReadJSONPtr(param, request.Params, &err)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	receipt, errC := this.pipe.Transactor().BroadcastTx(*param)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return receipt, 0, nil
}

func (this *ErisDbMethods) Transact(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &TransactParam{}
	err := this.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	receipt, errC := this.pipe.Transactor().Transact(param.PrivKey, param.Address, param.Data, param.GasLimit, param.Fee)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return receipt, 0, nil
}

func (this *ErisDbMethods) TransactAndHold(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &TransactParam{}
	err := this.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	ce, errC := this.pipe.Transactor().TransactAndHold(param.PrivKey, param.Address, param.Data, param.GasLimit, param.Fee)
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

func (this *ErisDbMethods) TransactNameReg(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &TransactNameRegParam{}
	err := this.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	receipt, errC := this.pipe.Transactor().TransactNameReg(param.PrivKey, param.Name, param.Data, param.Amount, param.Fee)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return receipt, 0, nil
}

func (this *ErisDbMethods) UnconfirmedTxs(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	txs, errC := this.pipe.Transactor().UnconfirmedTxs()
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return txs, 0, nil
}

func (this *ErisDbMethods) SignTx(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &SignTxParam{}
	err := this.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	tx := param.Tx
	pAccs := param.PrivAccounts
	txRet, errC := this.pipe.Transactor().SignTx(tx, pAccs)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return txRet, 0, nil
}

// *************************************** Name Registry ***************************************

func (this *ErisDbMethods) NameRegEntry(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &NameRegEntryParam{}
	err := this.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	name := param.Name
	// TODO is address check?
	entry, errC := this.pipe.NameReg().Entry(name)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return entry, 0, nil
}

func (this *ErisDbMethods) NameRegEntries(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &FilterListParam{}
	err := this.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	list, errC := this.pipe.NameReg().Entries(param.Filters)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return list, 0, nil
}

// **************************************************************************************

func generateSubId() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	rStr := hex.EncodeToString(b)
	return strings.ToUpper(rStr), nil

}
