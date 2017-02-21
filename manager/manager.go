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

package manager

import (
	"fmt"

	events "github.com/tendermint/go-events"

	config "github.com/eris-ltd/eris-db/config"
	definitions "github.com/eris-ltd/eris-db/definitions"
	erismint "github.com/eris-ltd/eris-db/manager/eris-mint"
	// types       "github.com/eris-ltd/eris-db/manager/types"

	"github.com/eris-ltd/eris-db/logging"
	"github.com/eris-ltd/eris-db/logging/loggers"
)

// NewApplicationPipe returns an initialised Pipe interface
// based on the loaded module configuration file.
// NOTE: [ben] Currently we only have a single `generic` definition
// of an application.  It is feasible this will be insufficient to support
// different types of applications later down the line.
func NewApplicationPipe(moduleConfig *config.ModuleConfig,
	evsw events.EventSwitch, logger loggers.InfoTraceLogger,
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
