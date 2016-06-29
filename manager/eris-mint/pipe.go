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

  db                "github.com/tendermint/go-db"
  tendermint_events "github.com/tendermint/go-events"
	tendermint_types  "github.com/tendermint/tendermint/types"
  wire              "github.com/tendermint/go-wire"

  log "github.com/eris-ltd/eris-logger"

	account              "github.com/eris-ltd/eris-db/account"
	config               "github.com/eris-ltd/eris-db/config"
	definitions          "github.com/eris-ltd/eris-db/definitions"
	event                "github.com/eris-ltd/eris-db/event"
	manager_types        "github.com/eris-ltd/eris-db/manager/types"
	rpc_tendermint_types "github.com/eris-ltd/eris-db/rpc/tendermint/core/types"
	state                "github.com/eris-ltd/eris-db/manager/eris-mint/state"
	state_types          "github.com/eris-ltd/eris-db/manager/eris-mint/state/types"
	transaction          "github.com/eris-ltd/eris-db/txs"
)

type ErisMintPipe struct {
  erisMintState   *state.State
  eventSwitch     *tendermint_events.EventSwitch
  erisMint        *ErisMint
  // Pipe implementations
  accounts        *accounts
  blockchain      *blockchain
  consensus       *consensus
  events          event.EventEmitter
  namereg         *namereg
  network         *network
  transactor      *transactor
  // Consensus interface
  consensusEngine definitions.ConsensusEngine
	// Genesis cache
	genesisDoc      *state_types.GenesisDoc
	genesisState    *state.State
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
    "chainId": startedState.ChainID,
    "lastBlockHeight": startedState.LastBlockHeight,
    "lastBlockHash": startedState.LastBlockHash,
    }).Debug("Loaded state")
  // start the application
  erisMint := NewErisMint(startedState, eventSwitch)

  // NOTE: [ben] Set Host opens an RPC pipe to Tendermint;  this is a remnant
  // of the old Eris-DB / Tendermint and should be considered as an in-process
  // call when possible
  tendermintHost := moduleConfig.Config.GetString("tendermint_host")
  log.Debug(fmt.Sprintf("Starting ErisMint RPC client to Tendermint host on %s",
    tendermintHost))
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

  return &ErisMintPipe {
    erisMintState: startedState,
    eventSwitch:   eventSwitch,
    erisMint:      erisMint,
    accounts:      accounts,
    events:        events,
    namereg:       namereg,
    transactor:    transactor,
		network:       newNetwork(),
    consensus:     nil,
		genesisDoc:    genesisDoc,
		genesisState:  nil,
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
    return nil, nil, fmt.Errorf("Dababase backend %s is not supported by %s",
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
	return nil, fmt.Errorf("Unimplemented.")
}

func (pipe *ErisMintPipe) Genesis() (*rpc_tendermint_types.ResultGenesis, error) {
	return nil, fmt.Errorf("Unimplemented.")
}

// Accounts
func (pipe *ErisMintPipe) GetAccount(address []byte) (*rpc_tendermint_types.ResultGetAccount,
	error) {
	return nil, fmt.Errorf("Unimplemented.")
}

func (pipe *ErisMintPipe) ListAccounts() (*rpc_tendermint_types.ResultListAccounts, error) {
	return nil, fmt.Errorf("Unimplemented.")
}

func (pipe *ErisMintPipe) GetStorage(address, key []byte) (*rpc_tendermint_types.ResultGetStorage,
	error) {
	return nil, fmt.Errorf("Unimplemented.")
}

func (pipe *ErisMintPipe) DumpStorage(address []byte) (*rpc_tendermint_types.ResultDumpStorage,
	error) {
	return nil, fmt.Errorf("Unimplemented.")
}

// Call
func (pipe *ErisMintPipe) Call(fromAddres, toAddress, data []byte) (*rpc_tendermint_types.ResultCall,
	error) {
	return nil, fmt.Errorf("Unimplemented.")
}

func (pipe *ErisMintPipe) CallCode(fromAddress, code, data []byte) (*rpc_tendermint_types.ResultCall,
	error) {
	return nil, fmt.Errorf("Unimplemented.")
}

// TODO: [ben] deprecate as we should not allow unsafe behaviour
// where a user is allowed to send a private key over the wire,
// especially unencrypted.
func (pipe *ErisMintPipe) SignTransaction(transaction transaction.Tx,
	privAccounts []*account.PrivAccount) (*rpc_tendermint_types.ResultSignTx,
	error) {
	return nil, fmt.Errorf("Unimplemented.")
}

// Name registry
func (pipe *ErisMintPipe) GetName(name string) (*rpc_tendermint_types.ResultGetName, error) {
	return nil, fmt.Errorf("Unimplemented.")
}

func (pipe *ErisMintPipe) ListNames() (*rpc_tendermint_types.ResultListNames, error) {
	return nil, fmt.Errorf("Unimplemented.")
}

// Memory pool
func (pipe *ErisMintPipe) BroadcastTxAsync(transaction transaction.Tx) (*rpc_tendermint_types.ResultBroadcastTx,
	error) {
	return nil, fmt.Errorf("Unimplemented.")
}

func (pipe *ErisMintPipe) BroadcastTxSync(transaction transaction.Tx) (*rpc_tendermint_types.ResultBroadcastTx,
	error) {
	return nil, fmt.Errorf("Unimplemented.")
}
