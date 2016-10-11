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

package consensus

import (
	"fmt"

	config "github.com/eris-ltd/eris-db/config"
	tendermint "github.com/eris-ltd/eris-db/consensus/tendermint"
	definitions "github.com/eris-ltd/eris-db/definitions"
)

func LoadConsensusEngineInPipe(moduleConfig *config.ModuleConfig,
	pipe definitions.Pipe) error {
	switch moduleConfig.Name {
	case "tendermint":
		tendermint, err := tendermint.NewTendermint(moduleConfig,
			pipe.GetApplication())
		if err != nil {
			return fmt.Errorf("Failed to load Tendermint node: %v", err)
		}

		err = pipe.SetConsensusEngine(tendermint)
		if err != nil {
			return fmt.Errorf("Failed to load Tendermint in pipe as "+
				"ConsensusEngine: %v", err)
		}

		// For Tendermint we have a coupled Blockchain and ConsensusEngine
		// implementation, so load it at the same time as ConsensusEngine
		err = pipe.SetBlockchain(tendermint)
		if err != nil {
			return fmt.Errorf("Failed to load Tendermint in pipe as "+
				"Blockchain: %v", err)
		}
	}
	return nil
}
