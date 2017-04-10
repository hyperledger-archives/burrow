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

package burrowmint

import (
	"bytes"
	"fmt"

	abci_types "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
	db "github.com/tendermint/go-db"
	go_events "github.com/tendermint/go-events"
	wire "github.com/tendermint/go-wire"
	tm_types "github.com/tendermint/tendermint/types"

	"github.com/monax/burrow/account"
	blockchain_types "github.com/monax/burrow/blockchain/types"
	imath "github.com/monax/burrow/common/math/integral"
	"github.com/monax/burrow/config"
	consensus_types "github.com/monax/burrow/consensus/types"
	core_types "github.com/monax/burrow/core/types"
	"github.com/monax/burrow/definitions"
	edb_event "github.com/monax/burrow/event"
	genesis "github.com/monax/burrow/genesis"
	"github.com/monax/burrow/logging"
	logging_types "github.com/monax/burrow/logging/types"
	vm "github.com/monax/burrow/manager/burrow-mint/evm"
	"github.com/monax/burrow/manager/burrow-mint/state"
	manager_types "github.com/monax/burrow/manager/types"
	rpc_tm_types "github.com/monax/burrow/rpc/tendermint/core/types"
	"github.com/monax/burrow/txs"
	"github.com/monax/burrow/word256"
)

type burrowMintPipe struct {
	burrowMintState *state.State
	burrowMint      *BurrowMint
	// Pipe implementations
	accounts        *accounts
	blockchain      blockchain_types.Blockchain
	consensusEngine consensus_types.ConsensusEngine
	events          edb_event.EventEmitter
	namereg         *namereg
	transactor      *transactor
	// Genesis cache
	genesisDoc   *genesis.GenesisDoc
	genesisState *state.State
	logger       logging_types.InfoTraceLogger
}

// Interface type assertions
var _ definitions.Pipe = (*burrowMintPipe)(nil)

var _ definitions.TendermintPipe = (*burrowMintPipe)(nil)

func NewBurrowMintPipe(moduleConfig *config.ModuleConfig,
	eventSwitch go_events.EventSwitch,
	logger logging_types.InfoTraceLogger) (*burrowMintPipe, error) {

	startedState, genesisDoc, err := startState(moduleConfig.DataDir,
		moduleConfig.Config.GetString("db_backend"), moduleConfig.GenesisFile,
		moduleConfig.ChainId)
	if err != nil {
		return nil, fmt.Errorf("Failed to start state: %v", err)
	}
	logger = logging.WithScope(logger, "BurrowMintPipe")
	// assert ChainId matches genesis ChainId
	logging.InfoMsg(logger, "Loaded state",
		"chainId", startedState.ChainID,
		"lastBlockHeight", startedState.LastBlockHeight,
		"lastBlockHash", startedState.LastBlockHash)
	// start the application
	burrowMint := NewBurrowMint(startedState, eventSwitch, logger)

	// initialise the components of the pipe
	events := edb_event.NewEvents(eventSwitch, logger)
	accounts := newAccounts(burrowMint)
	namereg := newNameReg(burrowMint)

	pipe := &burrowMintPipe{
		burrowMintState: startedState,
		burrowMint:      burrowMint,
		accounts:        accounts,
		events:          events,
		namereg:         namereg,
		// We need to set transactor later since we are introducing a mutual dependency
		// NOTE: this will be cleaned up when the RPC is unified
		transactor: nil,
		// genesis cache
		genesisDoc:   genesisDoc,
		genesisState: nil,
		// consensus and blockchain should both be loaded into the pipe by a higher
		// authority - this is a sort of dependency injection pattern
		consensusEngine: nil,
		blockchain:      nil,
		logger:          logger,
	}

	// NOTE: [Silas]
	// This is something of a loopback, but seems like a nicer option than
	// transactor calling the Tendermint native RPC (as it was before),
	// or indeed calling this RPC over the wire given that we have direct access.
	//
	// We could just hand transactor a copy of Pipe, but doing it this way seems
	// like a reasonably minimal and flexible way of providing transactor with the
	// broadcast function it needs, without making it explicitly
	// aware of/depend on Pipe.
	transactor := newTransactor(moduleConfig.ChainId, eventSwitch, burrowMint,
		events,
		func(tx txs.Tx) error {
			_, err := pipe.BroadcastTxSync(tx)
			return err
		})

	pipe.transactor = transactor
	return pipe, nil
}

//------------------------------------------------------------------------------
// Start state

// Start state tries to load the existing state in the data directory;
// if an existing database can be loaded, it will validate that the
// chainId in the genesis of that loaded state matches the asserted chainId.
// If no state can be loaded, the JSON genesis file will be loaded into the
// state database as the zero state.
func startState(dataDir, backend, genesisFile, chainId string) (*state.State,
	*genesis.GenesisDoc, error) {
	// avoid Tendermints PanicSanity and return a clean error
	if backend != db.MemDBBackendStr &&
		backend != db.LevelDBBackendStr {
		return nil, nil, fmt.Errorf("Database backend %s is not supported by %s",
			backend, GetBurrowMintVersion)
	}

	stateDB := db.NewDB("burrowmint", backend, dataDir)
	newState := state.LoadState(stateDB)
	var genesisDoc *genesis.GenesisDoc
	if newState == nil {
		genesisDoc, newState = state.MakeGenesisStateFromFile(stateDB, genesisFile)
		newState.Save()
		buf, n, err := new(bytes.Buffer), new(int), new(error)
		wire.WriteJSON(genesisDoc, buf, n, err)
		stateDB.Set(genesis.GenDocKey, buf.Bytes())
		if *err != nil {
			return nil, nil, fmt.Errorf("Unable to write genesisDoc to db: %v", err)
		}
	} else {
		loadedGenesisDocBytes := stateDB.Get(genesis.GenDocKey)
		err := new(error)
		wire.ReadJSONPtr(&genesisDoc, loadedGenesisDocBytes, err)
		if *err != nil {
			return nil, nil, fmt.Errorf("Unable to read genesisDoc from db on startState: %v", err)
		}
		// assert loaded genesis doc has the same chainId as the provided chainId
		if genesisDoc.ChainID != chainId {
			return nil, nil, fmt.Errorf("ChainId (%s) loaded from genesis document in existing database does not match"+
				" configuration chainId (%s).", genesisDoc.ChainID, chainId)
		}
	}

	return newState, genesisDoc, nil
}

//------------------------------------------------------------------------------
// Implement definitions.Pipe for burrowMintPipe

func (pipe *burrowMintPipe) Logger() logging_types.InfoTraceLogger {
	return pipe.logger
}

func (pipe *burrowMintPipe) Accounts() definitions.Accounts {
	return pipe.accounts
}

func (pipe *burrowMintPipe) Blockchain() blockchain_types.Blockchain {
	return pipe.blockchain
}

func (pipe *burrowMintPipe) Events() edb_event.EventEmitter {
	return pipe.events
}

func (pipe *burrowMintPipe) NameReg() definitions.NameReg {
	return pipe.namereg
}

func (pipe *burrowMintPipe) Transactor() definitions.Transactor {
	return pipe.transactor
}

func (pipe *burrowMintPipe) GetApplication() manager_types.Application {
	return pipe.burrowMint
}

func (pipe *burrowMintPipe) SetBlockchain(
	blockchain blockchain_types.Blockchain) error {
	if pipe.blockchain == nil {
		pipe.blockchain = blockchain
	} else {
		return fmt.Errorf("Failed to set Blockchain for pipe; already set")
	}
	return nil
}

func (pipe *burrowMintPipe) GetBlockchain() blockchain_types.Blockchain {
	return pipe.blockchain
}

func (pipe *burrowMintPipe) SetConsensusEngine(
	consensusEngine consensus_types.ConsensusEngine) error {
	if pipe.consensusEngine == nil {
		pipe.consensusEngine = consensusEngine
	} else {
		return fmt.Errorf("Failed to set consensus engine for pipe; already set")
	}
	return nil
}

func (pipe *burrowMintPipe) GetConsensusEngine() consensus_types.ConsensusEngine {
	return pipe.consensusEngine
}

func (pipe *burrowMintPipe) GetTendermintPipe() (definitions.TendermintPipe,
	error) {
	return definitions.TendermintPipe(pipe), nil
}

func (pipe *burrowMintPipe) consensusAndManagerEvents() edb_event.EventEmitter {
	// NOTE: [Silas] We could initialise this lazily and use the cached instance,
	// but for the time being that feels like a premature optimisation
	return edb_event.Multiplex(pipe.events, pipe.consensusEngine.Events())
}

//------------------------------------------------------------------------------
// Implement definitions.TendermintPipe for burrowMintPipe
func (pipe *burrowMintPipe) Subscribe(event string,
	rpcResponseWriter func(result rpc_tm_types.BurrowResult)) (*rpc_tm_types.ResultSubscribe, error) {
	subscriptionId, err := edb_event.GenerateSubId()
	if err != nil {
		return nil, err
		logging.InfoMsg(pipe.logger, "Subscribing to event",
			"event", event, "subscriptionId", subscriptionId)
	}
	pipe.consensusAndManagerEvents().Subscribe(subscriptionId, event,
		func(eventData txs.EventData) {
			result := rpc_tm_types.BurrowResult(&rpc_tm_types.ResultEvent{event,
				txs.EventData(eventData)})
			// NOTE: EventSwitch callbacks must be nonblocking
			rpcResponseWriter(result)
		})
	return &rpc_tm_types.ResultSubscribe{
		SubscriptionId: subscriptionId,
		Event:          event,
	}, nil
}

func (pipe *burrowMintPipe) Unsubscribe(subscriptionId string) (*rpc_tm_types.ResultUnsubscribe, error) {
	logging.InfoMsg(pipe.logger, "Unsubscribing from event",
		"subscriptionId", subscriptionId)
	pipe.consensusAndManagerEvents().Unsubscribe(subscriptionId)
	return &rpc_tm_types.ResultUnsubscribe{SubscriptionId: subscriptionId}, nil
}
func (pipe *burrowMintPipe) GenesisState() *state.State {
	if pipe.genesisState == nil {
		memoryDatabase := db.NewMemDB()
		pipe.genesisState = state.MakeGenesisState(memoryDatabase, pipe.genesisDoc)
	}
	return pipe.genesisState
}

func (pipe *burrowMintPipe) GenesisHash() []byte {
	return pipe.GenesisState().Hash()
}

func (pipe *burrowMintPipe) Status() (*rpc_tm_types.ResultStatus, error) {
	if pipe.consensusEngine == nil {
		return nil, fmt.Errorf("Consensus Engine not initialised in burrowmint pipe.")
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

func (pipe *burrowMintPipe) ChainId() (*rpc_tm_types.ResultChainId, error) {
	if pipe.blockchain == nil {
		return nil, fmt.Errorf("Blockchain not initialised in burrowmint pipe.")
	}
	chainId := pipe.blockchain.ChainId()

	return &rpc_tm_types.ResultChainId{
		ChainName:   chainId, // MARMOT: copy ChainId for ChainName as a placehodlder
		ChainId:     chainId,
		GenesisHash: pipe.GenesisHash(),
	}, nil
}

func (pipe *burrowMintPipe) NetInfo() (*rpc_tm_types.ResultNetInfo, error) {
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

func (pipe *burrowMintPipe) Genesis() (*rpc_tm_types.ResultGenesis, error) {
	return &rpc_tm_types.ResultGenesis{
		// TODO: [ben] sharing pointer to unmutated GenesisDoc, but is not immutable
		Genesis: pipe.genesisDoc,
	}, nil
}

// Accounts
func (pipe *burrowMintPipe) GetAccount(address []byte) (*rpc_tm_types.ResultGetAccount,
	error) {
	cache := pipe.burrowMint.GetCheckCache()
	account := cache.GetAccount(address)
	return &rpc_tm_types.ResultGetAccount{Account: account}, nil
}

func (pipe *burrowMintPipe) ListAccounts() (*rpc_tm_types.ResultListAccounts, error) {
	var blockHeight int
	var accounts []*account.Account
	state := pipe.burrowMint.GetState()
	blockHeight = state.LastBlockHeight
	state.GetAccounts().Iterate(func(key []byte, value []byte) bool {
		accounts = append(accounts, account.DecodeAccount(value))
		return false
	})
	return &rpc_tm_types.ResultListAccounts{blockHeight, accounts}, nil
}

func (pipe *burrowMintPipe) GetStorage(address, key []byte) (*rpc_tm_types.ResultGetStorage,
	error) {
	state := pipe.burrowMint.GetState()
	// state := consensusState.GetState()
	account := state.GetAccount(address)
	if account == nil {
		return nil, fmt.Errorf("UnknownAddress: %X", address)
	}
	storageRoot := account.StorageRoot
	storageTree := state.LoadStorage(storageRoot)

	_, value, exists := storageTree.Get(
		word256.LeftPadWord256(key).Bytes())
	if !exists {
		// value == nil {
		return &rpc_tm_types.ResultGetStorage{key, nil}, nil
	}
	return &rpc_tm_types.ResultGetStorage{key, value}, nil
}

func (pipe *burrowMintPipe) DumpStorage(address []byte) (*rpc_tm_types.ResultDumpStorage,
	error) {
	state := pipe.burrowMint.GetState()
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
func (pipe *burrowMintPipe) Call(fromAddress, toAddress, data []byte) (*rpc_tm_types.ResultCall,
	error) {
	st := pipe.burrowMint.GetState()
	cache := state.NewBlockCache(st)
	outAcc := cache.GetAccount(toAddress)
	if outAcc == nil {
		return nil, fmt.Errorf("Account %x does not exist", toAddress)
	}
	if fromAddress == nil {
		fromAddress = []byte{}
	}
	callee := toVMAccount(outAcc)
	caller := &vm.Account{Address: word256.LeftPadWord256(fromAddress)}
	txCache := state.NewTxCache(cache)
	gasLimit := st.GetGasLimit()
	params := vm.Params{
		BlockHeight: int64(st.LastBlockHeight),
		BlockHash:   word256.LeftPadWord256(st.LastBlockHash),
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

func (pipe *burrowMintPipe) CallCode(fromAddress, code, data []byte) (*rpc_tm_types.ResultCall,
	error) {
	st := pipe.burrowMint.GetState()
	cache := pipe.burrowMint.GetCheckCache()
	callee := &vm.Account{Address: word256.LeftPadWord256(fromAddress)}
	caller := &vm.Account{Address: word256.LeftPadWord256(fromAddress)}
	txCache := state.NewTxCache(cache)
	gasLimit := st.GetGasLimit()
	params := vm.Params{
		BlockHeight: int64(st.LastBlockHeight),
		BlockHash:   word256.LeftPadWord256(st.LastBlockHash),
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
func (pipe *burrowMintPipe) SignTransaction(tx txs.Tx,
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
func (pipe *burrowMintPipe) GetName(name string) (*rpc_tm_types.ResultGetName, error) {
	currentState := pipe.burrowMint.GetState()
	entry := currentState.GetNameRegEntry(name)
	if entry == nil {
		return nil, fmt.Errorf("Name %s not found", name)
	}
	return &rpc_tm_types.ResultGetName{entry}, nil
}

func (pipe *burrowMintPipe) ListNames() (*rpc_tm_types.ResultListNames, error) {
	var blockHeight int
	var names []*core_types.NameRegEntry
	currentState := pipe.burrowMint.GetState()
	blockHeight = currentState.LastBlockHeight
	currentState.GetNames().Iterate(func(key []byte, value []byte) bool {
		names = append(names, state.DecodeNameRegEntry(value))
		return false
	})
	return &rpc_tm_types.ResultListNames{blockHeight, names}, nil
}

func (pipe *burrowMintPipe) broadcastTx(tx txs.Tx,
	callback func(res *abci_types.Response)) (*rpc_tm_types.ResultBroadcastTx, error) {

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
func (pipe *burrowMintPipe) BroadcastTxAsync(tx txs.Tx) (*rpc_tm_types.ResultBroadcastTx, error) {
	return pipe.broadcastTx(tx, nil)
}

func (pipe *burrowMintPipe) BroadcastTxSync(tx txs.Tx) (*rpc_tm_types.ResultBroadcastTx, error) {
	responseChannel := make(chan *abci_types.Response, 1)
	_, err := pipe.broadcastTx(tx,
		func(res *abci_types.Response) {
			responseChannel <- res
		})
	if err != nil {
		return nil, err
	}
	// NOTE: [ben] This Response is set in /consensus/tendermint/local_client.go
	// a call to Application, here implemented by BurrowMint, over local callback,
	// or abci RPC call.  Hence the result is determined by BurrowMint/burrowmint.go
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
	case abci_types.CodeType_OK:
		return resultBroadCastTx, nil
	case abci_types.CodeType_EncodingError:
		return resultBroadCastTx, fmt.Errorf(resultBroadCastTx.Log)
	case abci_types.CodeType_InternalError:
		return resultBroadCastTx, fmt.Errorf(resultBroadCastTx.Log)
	default:
		logging.InfoMsg(pipe.logger, "Unknown error returned from Tendermint CheckTx on BroadcastTxSync",
			"application", GetBurrowMintVersion().GetVersionString(),
			"abci_code_type", responseCheckTx.Code,
			"abci_log", responseCheckTx.Log,
		)
		return resultBroadCastTx, fmt.Errorf("Unknown error returned: " + responseCheckTx.Log)
	}
}

func (pipe *burrowMintPipe) ListUnconfirmedTxs(maxTxs int) (*rpc_tm_types.ResultListUnconfirmedTxs, error) {
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
func (pipe *burrowMintPipe) BlockchainInfo(minHeight, maxHeight,
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

func (pipe *burrowMintPipe) GetBlock(height int) (*rpc_tm_types.ResultGetBlock, error) {
	return &rpc_tm_types.ResultGetBlock{
		Block:     pipe.blockchain.Block(height),
		BlockMeta: pipe.blockchain.BlockMeta(height),
	}, nil
}

func (pipe *burrowMintPipe) ListValidators() (*rpc_tm_types.ResultListValidators, error) {
	validators := pipe.consensusEngine.ListValidators()
	consensusState := pipe.consensusEngine.ConsensusState()
	// TODO: when we reintroduce support for bonding and unbonding update this
	// to reflect the mutable bonding state
	return &rpc_tm_types.ResultListValidators{
		BlockHeight:         consensusState.Height,
		BondedValidators:    validators,
		UnbondingValidators: nil,
	}, nil
}

func (pipe *burrowMintPipe) DumpConsensusState() (*rpc_tm_types.ResultDumpConsensusState, error) {
	statesMap := pipe.consensusEngine.PeerConsensusStates()
	peerStates := make([]*rpc_tm_types.ResultPeerConsensusState, len(statesMap))
	for key, peerState := range statesMap {
		peerStates = append(peerStates, &rpc_tm_types.ResultPeerConsensusState{
			PeerKey:            key,
			PeerConsensusState: peerState,
		})
	}
	dump := rpc_tm_types.ResultDumpConsensusState{
		ConsensusState:      pipe.consensusEngine.ConsensusState(),
		PeerConsensusStates: peerStates,
	}
	return &dump, nil
}
