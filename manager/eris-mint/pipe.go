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

  db     "github.com/tendermint/go-db"
  tendermint_events "github.com/tendermint/go-events"
  wire   "github.com/tendermint/go-wire"

  log "github.com/eris-ltd/eris-logger"

  config               "github.com/eris-ltd/eris-db/config"
  definitions          "github.com/eris-ltd/eris-db/definitions"
	event                "github.com/eris-ltd/eris-db/event"
  manager_types        "github.com/eris-ltd/eris-db/manager/types"
	// rpc_tendermint_types "github.com/eris-ltd/eris-db/rpc/tendermint/core/types"
  state                "github.com/eris-ltd/eris-db/manager/eris-mint/state"
  state_types          "github.com/eris-ltd/eris-db/manager/eris-mint/state/types"
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
  net             *network
  transactor      *transactor
  // Consensus interface
  consensusEngine definitions.ConsensusEngine
}

// NOTE [ben] Compiler check to ensure ErisMintPipe successfully implements
// eris-db/definitions.Pipe
var _ definitions.Pipe = (*ErisMintPipe)(nil)

// NOTE [ben] Compiler check to ensure ErisMintPipe successfully implements
// eris-db/definitions.TendermintPipe
// var _ definitions.TendermintPipe = (*ErisMintPipe)(nil)

func NewErisMintPipe(moduleConfig *config.ModuleConfig,
  eventSwitch *tendermint_events.EventSwitch) (*ErisMintPipe, error) {

  startedState, err := startState(moduleConfig.DataDir,
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
    consensus:     nil,
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
  error) {
  // avoid Tendermints PanicSanity and return a clean error
  if backend != db.DBBackendMemDB &&
    backend != db.DBBackendLevelDB {
    return nil, fmt.Errorf("Dababase backend %s is not supported by %s",
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
			return nil, fmt.Errorf("Unable to write genesisDoc to db: %v", err)
		}
	} else {
		loadedGenesisDocBytes := stateDB.Get(state_types.GenDocKey)
		err := new(error)
		wire.ReadJSONPtr(&genesisDoc, loadedGenesisDocBytes, err)
		if *err != nil {
			return nil, fmt.Errorf("Unable to read genesisDoc from db on startState: %v", err)
		}
    // assert loaded genesis doc has the same chainId as the provided chainId
    if genesisDoc.ChainID != chainId {
      return nil, fmt.Errorf("ChainId (%s) loaded from genesis document in existing database does not match configuration chainId (%s).",
      genesisDoc.ChainID, chainId)
    }
	}

  return newState, nil
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
  return pipe.net
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

//------------------------------------------------------------------------------
// Implement definitions.TendermintPipe for ErisMintPipe

// func (pipe *ErisMintPipe) Status() (*rpc_tendermint_types.ResultStatus, error) {
//
// }
