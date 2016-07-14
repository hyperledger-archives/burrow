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

	tendermint_common "github.com/tendermint/go-common"
	crypto "github.com/tendermint/go-crypto"
	db "github.com/tendermint/go-db"
	tendermint_events "github.com/tendermint/go-events"
	wire "github.com/tendermint/go-wire"
	tendermint_types "github.com/tendermint/tendermint/types"
	tmsp_types "github.com/tendermint/tmsp/types"

	log "github.com/eris-ltd/eris-logger"

	account "github.com/eris-ltd/eris-db/account"
	config "github.com/eris-ltd/eris-db/config"
	definitions "github.com/eris-ltd/eris-db/definitions"
	event "github.com/eris-ltd/eris-db/event"
	vm "github.com/eris-ltd/eris-db/manager/eris-mint/evm"
	state "github.com/eris-ltd/eris-db/manager/eris-mint/state"
	state_types "github.com/eris-ltd/eris-db/manager/eris-mint/state/types"
	manager_types "github.com/eris-ltd/eris-db/manager/types"
	rpc_tendermint_types "github.com/eris-ltd/eris-db/rpc/tendermint/core/types"
	core_types "github.com/eris-ltd/eris-db/core/types"
	"github.com/eris-ltd/eris-db/txs"
)

type ErisMintPipe struct {
	erisMintState *state.State
	eventSwitch   *tendermint_events.EventSwitch
	erisMint      *ErisMint
	// Pipe implementations
	accounts   *accounts
	blockchain *blockchain
	consensus  *consensus
	events     event.EventEmitter
	namereg    *namereg
	network    *network
	transactor *transactor
	// Consensus interface
	consensusEngine definitions.ConsensusEngine
	// Genesis cache
	genesisDoc   *state_types.GenesisDoc
	genesisState *state.State
}

// NOTE [ben] Compiler check to ensure ErisMintPipe successfully implements
// eris-db/definitions.Pipe
var _ definitions.Pipe = (*ErisMintPipe)(nil)

// NOTE [ben] Compiler check to ensure ErisMintPipe successfully implements
// eris-db/definitions.erisTendermintPipe
var _ definitions.TendermintPipe = (*ErisMintPipe)(nil)

func NewErisMintPipe(moduleConfig *config.ModuleConfig,
	eventSwitch *tendermint_events.EventSwitch) (*ErisMintPipe, error) {

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
	events := newEvents(eventSwitch)
	accounts := newAccounts(erisMint)
	namereg := newNameReg(erisMint)
	transactor := newTransactor(moduleConfig.ChainId, eventSwitch, erisMint,
		events)
	// TODO: make interface to tendermint core's rpc for these
	// blockchain := newBlockchain(chainID, genDocFile, blockStore)
	// consensus := newConsensus(erisdbApp)
	// net := newNetwork(erisdbApp)

	return &ErisMintPipe{
		erisMintState: startedState,
		eventSwitch:   eventSwitch,
		erisMint:      erisMint,
		accounts:      accounts,
		events:        events,
		namereg:       namereg,
		transactor:    transactor,
		network:       newNetwork(),
		consensus:     nil,
		// genesis cache
		genesisDoc:   genesisDoc,
		genesisState: nil,
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
			return nil, nil, fmt.Errorf("ChainId (%s) loaded from genesis document in existing database does not match configuration chainId (%s).",
				genesisDoc.ChainID, chainId)
		}
	}

	return newState, genesisDoc, nil
}

//------------------------------------------------------------------------------
// Implement definitions.Pipe for ErisMintPipe

func (pipe *ErisMintPipe) Accounts() definitions.Accounts {
	return pipe.accounts
}

func (pipe *ErisMintPipe) Blockchain() definitions.Blockchain {
	return pipe.blockchain
}

func (pipe *ErisMintPipe) Consensus() definitions.Consensus {
	return pipe.consensus
}

func (pipe *ErisMintPipe) Events() event.EventEmitter {
	return pipe.events
}

func (pipe *ErisMintPipe) NameReg() definitions.NameReg {
	return pipe.namereg
}

func (pipe *ErisMintPipe) Net() definitions.Net {
	return pipe.network
}

func (pipe *ErisMintPipe) Transactor() definitions.Transactor {
	return pipe.transactor
}

func (pipe *ErisMintPipe) GetApplication() manager_types.Application {
	return pipe.erisMint
}

func (pipe *ErisMintPipe) SetConsensusEngine(
	consensus definitions.ConsensusEngine) error {
	if pipe.consensusEngine == nil {
		pipe.consensusEngine = consensus
	} else {
		return fmt.Errorf("Failed to set consensus engine for pipe; already set")
	}
	return nil
}

func (pipe *ErisMintPipe) GetConsensusEngine() definitions.ConsensusEngine {
	return pipe.consensusEngine
}

func (pipe *ErisMintPipe) GetTendermintPipe() (definitions.TendermintPipe,
	error) {
	return definitions.TendermintPipe(pipe), nil
}

//------------------------------------------------------------------------------
// Implement definitions.TendermintPipe for ErisMintPipe

func (pipe *ErisMintPipe) Status() (*rpc_tendermint_types.ResultStatus, error) {
	memoryDatabase := db.NewMemDB()
	if pipe.genesisState == nil {
		pipe.genesisState = state.MakeGenesisState(memoryDatabase, pipe.genesisDoc)
	}
	genesisHash := pipe.genesisState.Hash()
	if pipe.consensusEngine == nil {
		return nil, fmt.Errorf("Consensus Engine is not set in pipe.")
	}
	latestHeight := pipe.consensusEngine.Height()
	var (
		latestBlockMeta *tendermint_types.BlockMeta
		latestBlockHash []byte
		latestBlockTime int64
	)
	if latestHeight != 0 {
		latestBlockMeta = pipe.consensusEngine.LoadBlockMeta(latestHeight)
		latestBlockHash = latestBlockMeta.Hash
		latestBlockTime = latestBlockMeta.Header.Time.UnixNano()
	}
	return &rpc_tendermint_types.ResultStatus{
		NodeInfo:          pipe.consensusEngine.NodeInfo(),
		GenesisHash:       genesisHash,
		PubKey:            pipe.consensusEngine.PublicValidatorKey(),
		LatestBlockHash:   latestBlockHash,
		LatestBlockHeight: latestHeight,
		LatestBlockTime:   latestBlockTime}, nil
}

func (pipe *ErisMintPipe) NetInfo() (*rpc_tendermint_types.ResultNetInfo, error) {
	listening := pipe.consensusEngine.IsListening()
	listeners := []string{}
	for _, listener := range pipe.consensusEngine.Listeners() {
		listeners = append(listeners, listener.String())
	}
	peers := pipe.consensusEngine.Peers()
	return &rpc_tendermint_types.ResultNetInfo{
		Listening: listening,
		Listeners: listeners,
		Peers:     peers,
	}, nil
}

func (pipe *ErisMintPipe) Genesis() (*rpc_tendermint_types.ResultGenesis, error) {
	return &rpc_tendermint_types.ResultGenesis{
		// TODO: [ben] sharing pointer to unmutated GenesisDoc, but is not immutable
		Genesis: pipe.genesisDoc,
	}, nil
}

// Accounts
func (pipe *ErisMintPipe) GetAccount(address []byte) (*rpc_tendermint_types.ResultGetAccount,
	error) {
	cache := pipe.erisMint.GetCheckCache()
	// cache := mempoolReactor.Mempool.GetCache()
	account := cache.GetAccount(address)
	if account == nil {
		log.Warn("Nil Account")
		return &rpc_tendermint_types.ResultGetAccount{nil}, nil
	}
	return &rpc_tendermint_types.ResultGetAccount{account}, nil
}

func (pipe *ErisMintPipe) ListAccounts() (*rpc_tendermint_types.ResultListAccounts, error) {
	var blockHeight int
	var accounts []*account.Account
	state := pipe.erisMint.GetState()
	blockHeight = state.LastBlockHeight
	state.GetAccounts().Iterate(func(key []byte, value []byte) bool {
		accounts = append(accounts, account.DecodeAccount(value))
		return false
	})
	return &rpc_tendermint_types.ResultListAccounts{blockHeight, accounts}, nil
}

func (pipe *ErisMintPipe) GetStorage(address, key []byte) (*rpc_tendermint_types.ResultGetStorage,
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
		tendermint_common.LeftPadWord256(key).Bytes())
	if !exists { // value == nil {
		return &rpc_tendermint_types.ResultGetStorage{key, nil}, nil
	}
	return &rpc_tendermint_types.ResultGetStorage{key, value}, nil
}

func (pipe *ErisMintPipe) DumpStorage(address []byte) (*rpc_tendermint_types.ResultDumpStorage,
	error) {
	state := pipe.erisMint.GetState()
	account := state.GetAccount(address)
	if account == nil {
		return nil, fmt.Errorf("UnknownAddress: %X", address)
	}
	storageRoot := account.StorageRoot
	storageTree := state.LoadStorage(storageRoot)
	storageItems := []rpc_tendermint_types.StorageItem{}
	storageTree.Iterate(func(key []byte, value []byte) bool {
		storageItems = append(storageItems, rpc_tendermint_types.StorageItem{key,
			value})
		return false
	})
	return &rpc_tendermint_types.ResultDumpStorage{storageRoot, storageItems}, nil
}

// Call
func (pipe *ErisMintPipe) Call(fromAddress, toAddress, data []byte) (*rpc_tendermint_types.ResultCall,
	error) {
	st := pipe.erisMint.GetState()
	cache := state.NewBlockCache(st)
	outAcc := cache.GetAccount(toAddress)
	if outAcc == nil {
		return nil, fmt.Errorf("Account %x does not exist", toAddress)
	}
	callee := toVMAccount(outAcc)
	caller := &vm.Account{Address: tendermint_common.LeftPadWord256(fromAddress)}
	txCache := state.NewTxCache(cache)
	params := vm.Params{
		BlockHeight: int64(st.LastBlockHeight),
		BlockHash:   tendermint_common.LeftPadWord256(st.LastBlockHash),
		BlockTime:   st.LastBlockTime.Unix(),
		GasLimit:    st.GetGasLimit(),
	}

	vmach := vm.NewVM(txCache, params, caller.Address, nil)
	gas := st.GetGasLimit()
	ret, err := vmach.Call(caller, callee, callee.Code, data, 0, &gas)
	if err != nil {
		return nil, err
	}
	return &rpc_tendermint_types.ResultCall{Return: ret}, nil
}

func (pipe *ErisMintPipe) CallCode(fromAddress, code, data []byte) (*rpc_tendermint_types.ResultCall,
	error) {
	st := pipe.erisMint.GetState()
	cache := pipe.erisMint.GetCheckCache()
	callee := &vm.Account{Address: tendermint_common.LeftPadWord256(fromAddress)}
	caller := &vm.Account{Address: tendermint_common.LeftPadWord256(fromAddress)}
	txCache := state.NewTxCache(cache)
	params := vm.Params{
		BlockHeight: int64(st.LastBlockHeight),
		BlockHash:   tendermint_common.LeftPadWord256(st.LastBlockHash),
		BlockTime:   st.LastBlockTime.Unix(),
		GasLimit:    st.GetGasLimit(),
	}

	vmach := vm.NewVM(txCache, params, caller.Address, nil)
	gas := st.GetGasLimit()
	ret, err := vmach.Call(caller, callee, code, data, 0, &gas)
	if err != nil {
		return nil, err
	}
	return &rpc_tendermint_types.ResultCall{Return: ret}, nil
}

// TODO: [ben] deprecate as we should not allow unsafe behaviour
// where a user is allowed to send a private key over the wire,
// especially unencrypted.
func (pipe *ErisMintPipe) SignTransaction(tx txs.Tx,
	privAccounts []*account.PrivAccount) (*rpc_tendermint_types.ResultSignTx,
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
	return &rpc_tendermint_types.ResultSignTx{tx}, nil
}

// Name registry
func (pipe *ErisMintPipe) GetName(name string) (*rpc_tendermint_types.ResultGetName, error) {
	currentState := pipe.erisMint.GetState()
	entry := currentState.GetNameRegEntry(name)
	if entry == nil {
		return nil, fmt.Errorf("Name %s not found", name)
	}
	return &rpc_tendermint_types.ResultGetName{entry}, nil
}

func (pipe *ErisMintPipe) ListNames() (*rpc_tendermint_types.ResultListNames, error) {
	var blockHeight int
	var names []*core_types.NameRegEntry
	currentState := pipe.erisMint.GetState()
	blockHeight = currentState.LastBlockHeight
	currentState.GetNames().Iterate(func(key []byte, value []byte) bool {
		names = append(names, state.DecodeNameRegEntry(value))
		return false
	})
	return &rpc_tendermint_types.ResultListNames{blockHeight, names}, nil
}

// Memory pool
// NOTE: txs must be signed
func (pipe *ErisMintPipe) BroadcastTxAsync(tx txs.Tx) (
	*rpc_tendermint_types.ResultBroadcastTx, error) {
	err := pipe.consensusEngine.BroadcastTransaction(txs.EncodeTx(tx), nil)
	if err != nil {
		return nil, fmt.Errorf("Error broadcasting txs: %v", err)
	}
	return &rpc_tendermint_types.ResultBroadcastTx{}, nil
}

func (pipe *ErisMintPipe) BroadcastTxSync(tx txs.Tx) (*rpc_tendermint_types.ResultBroadcastTx,
	error) {
	responseChannel := make(chan *tmsp_types.Response, 1)
	err := pipe.consensusEngine.BroadcastTransaction(txs.EncodeTx(tx),
		func(res *tmsp_types.Response) { responseChannel <- res })
	if err != nil {
		return nil, fmt.Errorf("Error broadcasting txs: %v", err)
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
	resultBroadCastTx := &rpc_tendermint_types.ResultBroadcastTx{
		Code: responseCheckTx.Code,
		Data: responseCheckTx.Data,
		Log:  responseCheckTx.Log,
	}
	fmt.Println("MARMOT resultBroadcastTx", resultBroadCastTx)
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
