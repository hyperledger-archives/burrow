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

package manager

import (
	"fmt"

	events "github.com/tendermint/go-events"

	config "github.com/eris-ltd/eris-db/config"
	definitions "github.com/eris-ltd/eris-db/definitions"
	logging_types "github.com/eris-ltd/eris-db/logging/types"
	erismint "github.com/eris-ltd/eris-db/manager/eris-mint"
	// types       "github.com/eris-ltd/eris-db/manager/types"

	"github.com/eris-ltd/eris-db/logging"
)

// NewApplicationPipe returns an initialised Pipe interface
// based on the loaded module configuration file.
// NOTE: [ben] Currently we only have a single `generic` definition
// of an application.  It is feasible this will be insufficient to support
// different types of applications later down the line.
func NewApplicationPipe(moduleConfig *config.ModuleConfig,
	evsw events.EventSwitch, logger logging_types.InfoTraceLogger,
	consensusMinorVersion string) (definitions.Pipe,
	error) {
	switch moduleConfig.Name {
	case "erismint":
		if err := erismint.AssertCompatibleConsensus(consensusMinorVersion); err != nil {
			return nil, err
		}
		logging.InfoMsg(logger, "Loading ErisMint",
			"compatibleConsensus", consensusMinorVersion,
			"erisMintVersion", erismint.GetErisMintVersion().GetVersionString())
		return erismint.NewErisMintPipe(moduleConfig, evsw, logger)
	}
	return nil, fmt.Errorf("Failed to return Pipe for %s", moduleConfig.Name)
}
