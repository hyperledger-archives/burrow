// Copyright 2015, 2016 Eris Industries (UK) Ltd.
// This file is part of Eris-RT

// Eris-RT is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Eris-RT is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Eris-RT.  If not, see <http://www.gnu.org/licenses/>.

package erismint

import (
	"bytes"
	"fmt"

	tm_common "github.com/tendermint/go-common"
	crypto "github.com/tendermint/go-crypto"
	db "github.com/tendermint/go-db"
	go_events "github.com/tendermint/go-events"
	wire "github.com/tendermint/go-wire"
	tm_types "github.com/tendermint/tendermint/types"
	tmsp_types "github.com/tendermint/tmsp/types"

	log "github.com/eris-ltd/eris-logger"

	"github.com/eris-ltd/eris-db/account"
	blockchain_types "github.com/eris-ltd/eris-db/blockchain/types"
	imath "github.com/eris-ltd/eris-db/common/math/integral"
	"github.com/eris-ltd/eris-db/config"
	consensus_types "github.com/eris-ltd/eris-db/consensus/types"
	core_types "github.com/eris-ltd/eris-db/core/types"
	"github.com/eris-ltd/eris-db/definitions"
	edb_event "github.com/eris-ltd/eris-db/event"
	vm "github.com/eris-ltd/eris-db/manager/eris-mint/evm"
	"github.com/eris-ltd/eris-db/manager/eris-mint/state"
	state_types "github.com/eris-ltd/eris-db/manager/eris-mint/state/types"
	manager_types "github.com/eris-ltd/eris-db/manager/types"
	rpc_tm_types "github.com/eris-ltd/eris-db/rpc/tendermint/core/types"
	"github.com/eris-ltd/eris-db/txs"
)

type erisMintPipe struct {
	erisMintState *state.State
	erisMint      *ErisMint
	// Pipe implementations
	accounts        *accounts
	blockchain      blockchain_types.Blockchain
	consensusEngine consensus_types.ConsensusEngine
	events          edb_event.EventEmitter
	namereg         *namereg
	network         *network
	transactor      *transactor
	// Genesis cache
	genesisDoc   *state_types.GenesisDoc
	genesisState *state.State
}

// NOTE [ben] Compiler check to ensure erisMintPipe successfully implements
// eris-db/definitions.Pipe
var _ definitions.Pipe = (*erisMintPipe)(nil)

// NOTE [ben] Compiler check to ensure erisMintPipe successfully implements
// eris-db/definitions.erisTendermintPipe
var _ definitions.TendermintPipe = (*erisMintPipe)(nil)

func NewErisMintPipe(moduleConfig *config.ModuleConfig,
	eventSwitch *go_events.EventSwitch) (*erisMintPipe, error) {

	startedState, genesisDoc, err := startState(moduleConfig.DataDir,
		moduleConfig.Config.GetString("db_backend"), moduleConfig.GenesisFile,
		moduleConfig.ChainId)
	if err != nil {
		return nil, fmt.Errorf("Failed to start state: %v", err)
	}
	// assert ChainId matches genesis ChainId
	log.WithFields(log.Fields{
		"chainId":         startedState.ChainID,
		"lastBlockHeight": startedState.LastBlockHeight,
		"lastBlockHash":   startedState.LastBlockHash,
	}).Debug("Loaded state")
	// start the application
	erisMint := NewErisMint(startedState, eventSwitch)

	// NOTE: [ben] Set Host opens an RPC pipe to Tendermint;  this is a remnant
	// of the old Eris-DB / Tendermint and should be considered as an in-process
	// call when possible
	tendermintHost := moduleConfig.Config.GetString("tendermint_host")
	erisMint.SetHostAddress(tendermintHost)

	// initialise the components of the pipe
	events := edb_event.NewEvents(eventSwitch)
	accounts := newAccounts(erisMint)
	namereg := newNameReg(erisMint)
	transactor := newTransactor(moduleConfig.ChainId, eventSwitch, erisMint,
		events)

	return &erisMintPipe{
		erisMintState: startedState,
		erisMint:      erisMint,
		accounts:      accounts,
		events:        events,
		namereg:       namereg,
		transactor:    transactor,
		// genesis cache
		genesisDoc:   genesisDoc,
		genesisState: nil,
		// TODO: What network-level information do we need?
		network: newNetwork(),
		// consensus and blockchain should both be loaded into the pipe by a higher
		// authority - this is a sort of dependency injection pattern
		consensusEngine: nil,
		blockchain:      nil,
	}, nil
}

//------------------------------------------------------------------------------
// Start state

// Start state tries to load the existing state in the data directory;
// if an existing database can be loaded, it will validate that the
// chainId in the genesis of that loaded state matches the asserted chainId.
// If no state can be loaded, the JSON genesis file will be loaded into the
// state database as the zero state.
func startState(dataDir, backend, genesisFile, chainId string) (*state.State,
	*state_types.GenesisDoc, error) {
	// avoid Tendermints PanicSanity and return a clean error
	if backend != db.DBBackendMemDB &&
		backend != db.DBBackendLevelDB {
		return nil, nil, fmt.Errorf("Database backend %s is not supported by %s",
			backend, GetErisMintVersion)
	}

	stateDB := db.NewDB("erismint", backend, dataDir)
	newState := state.LoadState(stateDB)
	var genesisDoc *state_types.GenesisDoc
	if newState == nil {
		genesisDoc, newState = state.MakeGenesisStateFromFile(stateDB, genesisFile)
		newState.Save()
		buf, n, err := new(bytes.Buffer), new(int), new(error)
		wire.WriteJSON(genesisDoc, buf, n, err)
		stateDB.Set(state_types.GenDocKey, buf.Bytes())
		if *err != nil {
			return nil, nil, fmt.Errorf("Unable to write genesisDoc to db: %v", err)
		}
	} else {
		loadedGenesisDocBytes := stateDB.Get(state_types.GenDocKey)
		err := new(error)
		wire.ReadJSONPtr(&genesisDoc, loadedGenesisDocBytes, err)
		if *err != nil {
			return nil, nil, fmt.Errorf("Unable to read genesisDoc from db on startState: %v", err)
		}
		// assert loaded genesis doc has the same chainId as the provided chainId
		if genesisDoc.ChainID != chainId {
			log.WithFields(log.Fields{
				"chainId from loaded genesis": genesisDoc.ChainID,
				"chainId from configuration":  chainId,
			}).Warn("Conflicting chainIds")
			// return nil, nil, fmt.Errorf("ChainId (%s) loaded from genesis document in existing database does not match configuration chainId (%s).",
			// genesisDoc.ChainID, chainId)
		}
	}

	return newState, genesisDoc, nil
}

//------------------------------------------------------------------------------
// Implement definitions.Pipe for erisMintPipe

func (pipe *erisMintPipe) Accounts() definitions.Accounts {
	return pipe.accounts
}

func (pipe *erisMintPipe) Blockchain() blockchain_types.Blockchain {
	return pipe.blockchain
}

func (pipe *erisMintPipe) Consensus() consensus_types.ConsensusEngine {
	return pipe.consensusEngine
}

func (pipe *erisMintPipe) Events() edb_event.EventEmitter {
	return pipe.events
}

func (pipe *erisMintPipe) NameReg() definitions.NameReg {
	return pipe.namereg
}

func (pipe *erisMintPipe) Net() definitions.Net {
	return pipe.network
}

func (pipe *erisMintPipe) Transactor() definitions.Transactor {
	return pipe.transactor
}

func (pipe *erisMintPipe) GetApplication() manager_types.Application {
	return pipe.erisMint
}

func (pipe *erisMintPipe) SetBlockchain(
	blockchain blockchain_types.Blockchain) error {
	if pipe.blockchain == nil {
		pipe.blockchain = blockchain
	} else {
		return fmt.Errorf("Failed to set Blockchain for pipe; already set")
	}
	return nil
}

func (pipe *erisMintPipe) GetBlockchain() blockchain_types.Blockchain {
	return pipe.blockchain
}

func (pipe *erisMintPipe) SetConsensusEngine(
	consensusEngine consensus_types.ConsensusEngine) error {
	if pipe.consensusEngine == nil {
		pipe.consensusEngine = consensusEngine
	} else {
		return fmt.Errorf("Failed to set consensus engine for pipe; already set")
	}
	return nil
}

func (pipe *erisMintPipe) GetConsensusEngine() consensus_types.ConsensusEngine {
	return pipe.consensusEngine
}

func (pipe *erisMintPipe) GetTendermintPipe() (definitions.TendermintPipe,
	error) {
	return definitions.TendermintPipe(pipe), nil
}

func (pipe *erisMintPipe) consensusAndManagerEvents() edb_event.EventEmitter {
	// NOTE: [Silas] We could initialise this lazily and use the cached instance,
	// but for the time being that feels like a premature optimisation
	return edb_event.Multiplex(pipe.events, pipe.consensusEngine.Events())
}

//------------------------------------------------------------------------------
// Implement definitions.TendermintPipe for erisMintPipe
func (pipe *erisMintPipe) Subscribe(event string,
	rpcResponseWriter func(result rpc_tm_types.ErisDBResult)) (*rpc_tm_types.ResultSubscribe, error) {
	subscriptionId, err := edb_event.GenerateSubId()
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{"event": event, "subscriptionId": subscriptionId}).
		Info("Subscribing to event")

	pipe.consensusAndManagerEvents().Subscribe(subscriptionId, event,
		func(eventData txs.EventData) {
			result := rpc_tm_types.ErisDBResult(&rpc_tm_types.ResultEvent{event,
				txs.EventData(eventData)})
			// NOTE: EventSwitch callbacks must be nonblocking
			rpcResponseWriter(result)
		})
	return &rpc_tm_types.ResultSubscribe{
		SubscriptionId: subscriptionId,
		Event:          event,
	}, nil
}

func (pipe *erisMintPipe) Unsubscribe(subscriptionId string) (*rpc_tm_types.ResultUnsubscribe, error) {
	log.WithFields(log.Fields{"subscriptionId": subscriptionId}).
		Info("Unsubscribing from event")
	pipe.consensusAndManagerEvents().Unsubscribe(subscriptionId)
	return &rpc_tm_types.ResultUnsubscribe{SubscriptionId: subscriptionId}, nil
}
func (pipe *erisMintPipe) GenesisState() *state.State {
	if pipe.genesisState == nil {
		memoryDatabase := db.NewMemDB()
		pipe.genesisState = state.MakeGenesisState(memoryDatabase, pipe.genesisDoc)
	}
	return pipe.genesisState
}

func (pipe *erisMintPipe) GenesisHash() []byte {
	return pipe.GenesisState().Hash()
}

func (pipe *erisMintPipe) Status() (*rpc_tm_types.ResultStatus, error) {
	if pipe.consensusEngine == nil {
		return nil, fmt.Errorf("Consensus Engine is not set in pipe.")
	}
	latestHeight := pipe.blockchain.Height()
	var (
		latestBlockMeta *tm_types.BlockMeta
		latestBlockHash []byte
		latestBlockTime int64
	)
	if latestHeight != 0 {
		latestBlockMeta = pipe.blockchain.BlockMeta(latestHeight)
		latestBlockHash = latestBlockMeta.Hash
		latestBlockTime = latestBlockMeta.Header.Time.UnixNano()
	}
	return &rpc_tm_types.ResultStatus{
		NodeInfo:          pipe.consensusEngine.NodeInfo(),
		GenesisHash:       pipe.GenesisHash(),
		PubKey:            pipe.consensusEngine.PublicValidatorKey(),
		LatestBlockHash:   latestBlockHash,
		LatestBlockHeight: latestHeight,
		LatestBlockTime:   latestBlockTime}, nil
}

func (pipe *erisMintPipe) NetInfo() (*rpc_tm_types.ResultNetInfo, error) {
	listening := pipe.consensusEngine.IsListening()
	listeners := []string{}
	for _, listener := range pipe.consensusEngine.Listeners() {
		listeners = append(listeners, listener.String())
	}
	peers := pipe.consensusEngine.Peers()
	return &rpc_tm_types.ResultNetInfo{
		Listening: listening,
		Listeners: listeners,
		Peers:     peers,
	}, nil
}

func (pipe *erisMintPipe) Genesis() (*rpc_tm_types.ResultGenesis, error) {
	return &rpc_tm_types.ResultGenesis{
		// TODO: [ben] sharing pointer to unmutated GenesisDoc, but is not immutable
		Genesis: pipe.genesisDoc,
	}, nil
}

// Accounts
func (pipe *erisMintPipe) GetAccount(address []byte) (*rpc_tm_types.ResultGetAccount,
	error) {
	cache := pipe.erisMint.GetCheckCache()
	// cache := mempoolReactor.Mempool.GetCache()
	account := cache.GetAccount(address)
	if account == nil {
		log.Warn("Nil Account")
		return &rpc_tm_types.ResultGetAccount{nil}, nil
	}
	return &rpc_tm_types.ResultGetAccount{account}, nil
}

func (pipe *erisMintPipe) ListAccounts() (*rpc_tm_types.ResultListAccounts, error) {
	var blockHeight int
	var accounts []*account.Account
	state := pipe.erisMint.GetState()
	blockHeight = state.LastBlockHeight
	state.GetAccounts().Iterate(func(key []byte, value []byte) bool {
		accounts = append(accounts, account.DecodeAccount(value))
		return false
	})
	return &rpc_tm_types.ResultListAccounts{blockHeight, accounts}, nil
}

func (pipe *erisMintPipe) GetStorage(address, key []byte) (*rpc_tm_types.ResultGetStorage,
	error) {
	state := pipe.erisMint.GetState()
	// state := consensusState.GetState()
	account := state.GetAccount(address)
	if account == nil {
		return nil, fmt.Errorf("UnknownAddress: %X", address)
	}
	storageRoot := account.StorageRoot
	storageTree := state.LoadStorage(storageRoot)

	_, value, exists := storageTree.Get(
		tm_common.LeftPadWord256(key).Bytes())
	if !exists {
		// value == nil {
		return &rpc_tm_types.ResultGetStorage{key, nil}, nil
	}
	return &rpc_tm_types.ResultGetStorage{key, value}, nil
}

func (pipe *erisMintPipe) DumpStorage(address []byte) (*rpc_tm_types.ResultDumpStorage,
	error) {
	state := pipe.erisMint.GetState()
	account := state.GetAccount(address)
	if account == nil {
		return nil, fmt.Errorf("UnknownAddress: %X", address)
	}
	storageRoot := account.StorageRoot
	storageTree := state.LoadStorage(storageRoot)
	storageItems := []rpc_tm_types.StorageItem{}
	storageTree.Iterate(func(key []byte, value []byte) bool {
		storageItems = append(storageItems, rpc_tm_types.StorageItem{key,
			value})
		return false
	})
	return &rpc_tm_types.ResultDumpStorage{storageRoot, storageItems}, nil
}

// Call 
// NOTE: this function is used from 46657 and has sibling on 1337
// in transactor.go
// TODO: [ben] resolve incompatibilities in byte representation for 0.12.0 release
func (pipe *erisMintPipe) Call(fromAddress, toAddress, data []byte) (*rpc_tm_types.ResultCall,
	error) {
	st := pipe.erisMint.GetState()
	cache := state.NewBlockCache(st)
	outAcc := cache.GetAccount(toAddress)
	if outAcc == nil {
		return nil, fmt.Errorf("Account %x does not exist", toAddress)
	}
	if fromAddress == nil {
		fromAddress = []byte{}
	}
	callee := toVMAccount(outAcc)
	caller := &vm.Account{Address: tm_common.LeftPadWord256(fromAddress)}
	txCache := state.NewTxCache(cache)
	gasLimit := st.GetGasLimit()
	params := vm.Params{
		BlockHeight: int64(st.LastBlockHeight),
		BlockHash:   tm_common.LeftPadWord256(st.LastBlockHash),
		BlockTime:   st.LastBlockTime.Unix(),
		GasLimit:    gasLimit,
	}

	vmach := vm.NewVM(txCache, params, caller.Address, nil)
	gas := gasLimit
	ret, err := vmach.Call(caller, callee, callee.Code, data, 0, &gas)
	if err != nil {
		return nil, err
	}
	gasUsed := gasLimit - gas
	// here return bytes are not hex encoded; on the sibling function
	// they are
	return &rpc_tm_types.ResultCall{Return: ret, GasUsed: gasUsed}, nil
}

func (pipe *erisMintPipe) CallCode(fromAddress, code, data []byte) (*rpc_tm_types.ResultCall,
	error) {
	st := pipe.erisMint.GetState()
	cache := pipe.erisMint.GetCheckCache()
	callee := &vm.Account{Address: tm_common.LeftPadWord256(fromAddress)}
	caller := &vm.Account{Address: tm_common.LeftPadWord256(fromAddress)}
	txCache := state.NewTxCache(cache)
	gasLimit := st.GetGasLimit()
	params := vm.Params{
		BlockHeight: int64(st.LastBlockHeight),
		BlockHash:   tm_common.LeftPadWord256(st.LastBlockHash),
		BlockTime:   st.LastBlockTime.Unix(),
		GasLimit:    gasLimit,
	}

	vmach := vm.NewVM(txCache, params, caller.Address, nil)
	gas := gasLimit
	ret, err := vmach.Call(caller, callee, code, data, 0, &gas)
	if err != nil {
		return nil, err
	}
	gasUsed := gasLimit - gas
	return &rpc_tm_types.ResultCall{Return: ret, GasUsed: gasUsed}, nil
}

// TODO: [ben] deprecate as we should not allow unsafe behaviour
// where a user is allowed to send a private key over the wire,
// especially unencrypted.
func (pipe *erisMintPipe) SignTransaction(tx txs.Tx,
	privAccounts []*account.PrivAccount) (*rpc_tm_types.ResultSignTx,
	error) {

	for i, privAccount := range privAccounts {
		if privAccount == nil || privAccount.PrivKey == nil {
			return nil, fmt.Errorf("Invalid (empty) privAccount @%v", i)
		}
	}
	switch tx.(type) {
	case *txs.SendTx:
		sendTx := tx.(*txs.SendTx)
		for i, input := range sendTx.Inputs {
			input.PubKey = privAccounts[i].PubKey
			input.Signature = privAccounts[i].Sign(pipe.transactor.chainID, sendTx)
		}
	case *txs.CallTx:
		callTx := tx.(*txs.CallTx)
		callTx.Input.PubKey = privAccounts[0].PubKey
		callTx.Input.Signature = privAccounts[0].Sign(pipe.transactor.chainID, callTx)
	case *txs.BondTx:
		bondTx := tx.(*txs.BondTx)
		// the first privaccount corresponds to the BondTx pub key.
		// the rest to the inputs
		bondTx.Signature = privAccounts[0].Sign(pipe.transactor.chainID, bondTx).(crypto.SignatureEd25519)
		for i, input := range bondTx.Inputs {
			input.PubKey = privAccounts[i+1].PubKey
			input.Signature = privAccounts[i+1].Sign(pipe.transactor.chainID, bondTx)
		}
	case *txs.UnbondTx:
		unbondTx := tx.(*txs.UnbondTx)
		unbondTx.Signature = privAccounts[0].Sign(pipe.transactor.chainID, unbondTx).(crypto.SignatureEd25519)
	case *txs.RebondTx:
		rebondTx := tx.(*txs.RebondTx)
		rebondTx.Signature = privAccounts[0].Sign(pipe.transactor.chainID, rebondTx).(crypto.SignatureEd25519)
	}
	return &rpc_tm_types.ResultSignTx{tx}, nil
}

// Name registry
func (pipe *erisMintPipe) GetName(name string) (*rpc_tm_types.ResultGetName, error) {
	currentState := pipe.erisMint.GetState()
	entry := currentState.GetNameRegEntry(name)
	if entry == nil {
		return nil, fmt.Errorf("Name %s not found", name)
	}
	return &rpc_tm_types.ResultGetName{entry}, nil
}

func (pipe *erisMintPipe) ListNames() (*rpc_tm_types.ResultListNames, error) {
	var blockHeight int
	var names []*core_types.NameRegEntry
	currentState := pipe.erisMint.GetState()
	blockHeight = currentState.LastBlockHeight
	currentState.GetNames().Iterate(func(key []byte, value []byte) bool {
		names = append(names, state.DecodeNameRegEntry(value))
		return false
	})
	return &rpc_tm_types.ResultListNames{blockHeight, names}, nil
}

func (pipe *erisMintPipe) broadcastTx(tx txs.Tx,
	callback func(res *tmsp_types.Response)) (*rpc_tm_types.ResultBroadcastTx, error) {

	txBytes, err := txs.EncodeTx(tx)
	if err != nil {
		return nil, fmt.Errorf("Error encoding transaction: %v", err)
	}
	err = pipe.consensusEngine.BroadcastTransaction(txBytes, callback)
	if err != nil {
		return nil, fmt.Errorf("Error broadcasting transaction: %v", err)
	}
	return &rpc_tm_types.ResultBroadcastTx{}, nil
}

// Memory pool
// NOTE: txs must be signed
func (pipe *erisMintPipe) BroadcastTxAsync(tx txs.Tx) (*rpc_tm_types.ResultBroadcastTx, error) {
	return pipe.broadcastTx(tx, nil)
}

func (pipe *erisMintPipe) BroadcastTxSync(tx txs.Tx) (*rpc_tm_types.ResultBroadcastTx, error) {
	responseChannel := make(chan *tmsp_types.Response, 1)
	_, err := pipe.broadcastTx(tx,
		func(res *tmsp_types.Response) {
			responseChannel <- res
		})
	if err != nil {
		return nil, err
	}
	// NOTE: [ben] This Response is set in /consensus/tendermint/local_client.go
	// a call to Application, here implemented by ErisMint, over local callback,
	// or TMSP RPC call.  Hence the result is determined by ErisMint/erismint.go
	// CheckTx() Result (Result converted to ReqRes into Response returned here)
	// NOTE: [ben] BroadcastTx just calls CheckTx in Tendermint (oddly... [Silas])
	response := <-responseChannel
	responseCheckTx := response.GetCheckTx()
	if responseCheckTx == nil {
		return nil, fmt.Errorf("Error, application did not return CheckTx response.")
	}
	resultBroadCastTx := &rpc_tm_types.ResultBroadcastTx{
		Code: responseCheckTx.Code,
		Data: responseCheckTx.Data,
		Log:  responseCheckTx.Log,
	}
	switch responseCheckTx.Code {
	case tmsp_types.CodeType_OK:
		return resultBroadCastTx, nil
	case tmsp_types.CodeType_EncodingError:
		return resultBroadCastTx, fmt.Errorf(resultBroadCastTx.Log)
	case tmsp_types.CodeType_InternalError:
		return resultBroadCastTx, fmt.Errorf(resultBroadCastTx.Log)
	default:
		log.WithFields(log.Fields{
			"application":    GetErisMintVersion().GetVersionString(),
			"TMSP_code_type": responseCheckTx.Code,
		}).Warn("Unknown error returned from Tendermint CheckTx on BroadcastTxSync")
		return resultBroadCastTx, fmt.Errorf("Unknown error returned: " + responseCheckTx.Log)
	}
}

func (pipe *erisMintPipe) ListUnconfirmedTxs(maxTxs int) (*rpc_tm_types.ResultListUnconfirmedTxs, error) {
	// Get all transactions for now
	transactions, err := pipe.consensusEngine.ListUnconfirmedTxs(maxTxs)
	if err != nil {
		return nil, err
	}
	return &rpc_tm_types.ResultListUnconfirmedTxs{
		N:   len(transactions),
		Txs: transactions,
	}, nil
}

// Returns the current blockchain height and metadata for a range of blocks
// between minHeight and maxHeight. Only returns maxBlockLookback block metadata
// from the top of the range of blocks.
// Passing 0 for maxHeight sets the upper height of the range to the current
// blockchain height.
func (pipe *erisMintPipe) BlockchainInfo(minHeight, maxHeight,
	maxBlockLookback int) (*rpc_tm_types.ResultBlockchainInfo, error) {

	latestHeight := pipe.blockchain.Height()

	if maxHeight < 1 {
		maxHeight = latestHeight
	} else {
		maxHeight = imath.MinInt(latestHeight, maxHeight)
	}
	if minHeight < 1 {
		minHeight = imath.MaxInt(1, maxHeight-maxBlockLookback)
	}

	blockMetas := []*tm_types.BlockMeta{}
	for height := maxHeight; height >= minHeight; height-- {
		blockMeta := pipe.blockchain.BlockMeta(height)
		blockMetas = append(blockMetas, blockMeta)
	}

	return &rpc_tm_types.ResultBlockchainInfo{
		LastHeight: latestHeight,
		BlockMetas: blockMetas,
	}, nil
}

func (pipe *erisMintPipe) GetBlock(height int) (*rpc_tm_types.ResultGetBlock, error) {
	return &rpc_tm_types.ResultGetBlock{
		Block:     pipe.blockchain.Block(height),
		BlockMeta: pipe.blockchain.BlockMeta(height),
	}, nil
}

func (pipe *erisMintPipe) ListValidators() (*rpc_tm_types.ResultListValidators, error) {
	validators := pipe.consensusEngine.ListValidators()
	consensusState := pipe.consensusEngine.ConsensusState()
	// TODO: when we reintroduce support for bonding and unbonding update this
	// to reflect the mutable bonding state
	return &rpc_tm_types.ResultListValidators{
		BlockHeight: consensusState.Height,
		BondedValidators: validators,
		UnbondingValidators: nil,
	}, nil
}

func (pipe *erisMintPipe) DumpConsensusState() (*rpc_tm_types.ResultDumpConsensusState, error) {
	return &rpc_tm_types.ResultDumpConsensusState{
		ConsensusState: pipe.consensusEngine.ConsensusState(),
		PeerConsensusStates: pipe.consensusEngine.PeerConsensusStates(),
	}, nil
}
