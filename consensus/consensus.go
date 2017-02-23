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
		tmint, err := tendermint.NewTendermint(moduleConfig, pipe.GetApplication(),
			pipe.Logger().With())
		if err != nil {
			return fmt.Errorf("Failed to load Tendermint node: %v", err)
		}

		err = pipe.SetConsensusEngine(tmint)
		if err != nil {
			return fmt.Errorf("Failed to load Tendermint in pipe as "+
				"ConsensusEngine: %v", err)
		}

		// For Tendermint we have a coupled Blockchain and ConsensusEngine
		// implementation, so load it at the same time as ConsensusEngine
		err = pipe.SetBlockchain(tmint)
		if err != nil {
			return fmt.Errorf("Failed to load Tendermint in pipe as "+
				"Blockchain: %v", err)
		}
	}
	return nil
}
