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

  db   "github.com/tendermint/go-db"
  wire "github.com/tendermint/go-wire"

  config      "github.com/eris-ltd/eris-db/config"
  state       "github.com/eris-ltd/eris-db/manager/eris-mint/state"
  state_types "github.com/eris-ltd/eris-db/manager/eris-mint/state/types"
)

type ErisMintPipe struct {
  erisMintState *state.State
}

func NewErisMintPipe(moduleConfig *config.ModuleConfig,
  genesisFile string) (*ErisMintPipe, error) {

  startedState, err := startState(moduleConfig.DataDir,
    moduleConfig.Config.GetString("db_backend"), genesisFile)
  if err != nil {
    return nil, fmt.Errorf("Failed to start state: %v", err)
  }
  return &ErisMintPipe{
    erisMintState: startedState,
  }, nil
}

//------------------------------------------------------------------------------
// Start state

func startState(dataDir, backend, genesisFile string) (*state.State, error) {
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
		genDocBytes := stateDB.Get(state_types.GenDocKey)
		err := new(error)
		wire.ReadJSONPtr(&genesisDoc, genDocBytes, err)
		if *err != nil {
			return nil, fmt.Errorf("Unable to read genesisDoc from db: %v", err)
		}
	}

  return newState, nil
}
