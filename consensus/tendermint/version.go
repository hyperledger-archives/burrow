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

package tendermint

import (
  version "github.com/eris-ltd/eris-db/version"
)

const (
  // Client identifier to advertise over the network
  tendermintClientIdentifier = "tendermint"
  // Major version component of the current release
  tendermintVersionMajor     = 0
  // Minor version component of the current release
  tendermintVersionMinor     = 6
  // Patch version component of the current release
  tendermintVersionPatch     = 0
)

func GetTendermintVersion() *version.VersionIdentifier {
  return version.New(tendermintClientIdentifier, tendermintVersionMajor,
    tendermintVersionMinor, tendermintVersionPatch)
}
