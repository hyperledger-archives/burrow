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

package core

import (
  "fmt"

  // TODO: [ben] swap out go-events with eris-db/event (currently unused)
  events "github.com/tendermint/go-events"

  config      "github.com/eris-ltd/eris-db/config"
  consensus   "github.com/eris-ltd/eris-db/consensus"
  definitions "github.com/eris-ltd/eris-db/definitions"
  manager     "github.com/eris-ltd/eris-db/manager"
)

// Core is the high-level structure
type Core struct {
  chainId string
  evsw    *events.EventSwitch
  pipe    definitions.Pipe
}

func NewCore(chainId, genesisFile string, consensusConfig *config.ModuleConfig,
  managerConfig *config.ModuleConfig) (*Core, error) {

  // start new event switch, TODO: [ben] replace with eris-db/event
  evsw := events.NewEventSwitch()
  evsw.Start()

  // start a new application pipe that will load an application manager
  pipe, err := manager.NewApplicationPipe(managerConfig, evsw,
    consensusConfig.Version, genesisFile)
  if err != nil {
    return nil, fmt.Errorf("PLACEHOLDER")
  }
  // pass the
  consensus.LoadConsensusEngineInPipe(consensusConfig, pipe)


  // create state
  // from genesis
  // create event switch
  // give state and evsw to app
  // give app to consensus
  // create new Pipe
  // give app
  // create servers
  return &Core{}, fmt.Errorf("PLACEHOLDER")
}

//------------------------------------------------------------------------------
// Explicit switch that can later be abstracted into an `Engine` definition
// where the Engine defines the explicit interaction of a specific application
// manager with a consensus engine.
// TODO: [ben] before such Engine abstraction,
// think about many-manager-to-one-consensus
