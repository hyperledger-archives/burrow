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

// version provides the current Eris-DB version and a VersionIdentifier
// for the modules to identify their version with.

package consensus

import (
  // noops      "github.com/eris-ltd/eris-db/consensus/noops"
  tendermint "github.com/eris-ltd/eris-db/consensus/tendermint"
  tmsp       "github.com/eris-ltd/eris-db/consensus/tmsp"
)

//------------------------------------------------------------------------------
// Helper functions

func AssertValidConsensusModule(name, minorVersionString string) bool {
  switch name {
  case "noops" :
    // noops should not have any external interfaces that can change
    // over iterations
    return true
  case "tmsp" :
    return minorVersionString == tmsp.GetTmspVersion().GetMinorVersionString()
  case "tendermint" :
    return minorVersionString == tendermint.GetTendermintVersion().GetMinorVersionString()
  }
  return false
}
