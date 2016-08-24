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
	"strconv"

	tendermint_version "github.com/tendermint/tendermint/version"

	version "github.com/eris-ltd/eris-db/version"
)

const (
	// Client identifier to advertise over the network
	tendermintClientIdentifier = "tendermint"
	// Major version component of the current release
	tendermintVersionMajorConst uint8 = 0
	// Minor version component of the current release
	tendermintVersionMinorConst uint8 = 6
	// Patch version component of the current release
	tendermintVersionPatchConst uint8 = 0
)

var (
	tendermintVersionMajor uint8
	tendermintVersionMinor uint8
	tendermintVersionPatch uint8
)

func init() {
	// discard error because we test for this in Continuous Integration tests
	tendermintVersionMajor, _ = getTendermintMajorVersionFromSource()
	tendermintVersionMinor, _ = getTendermintMinorVersionFromSource()
	tendermintVersionPatch, _ = getTendermintPatchVersionFromSource()
}

func GetTendermintVersion() *version.VersionIdentifier {
	return version.New(tendermintClientIdentifier, tendermintVersionMajor,
		tendermintVersionMinor, tendermintVersionPatch)
}

func getTendermintMajorVersionFromSource() (uint8, error) {
	majorVersionUint, err := strconv.ParseUint(tendermint_version.Maj, 10, 8)
	if err != nil {
		return tendermintVersionMajorConst, err
	}
	return uint8(majorVersionUint), nil
}

func getTendermintMinorVersionFromSource() (uint8, error) {
	minorVersionUint, err := strconv.ParseUint(tendermint_version.Min, 10, 8)
	if err != nil {
		return tendermintVersionMinorConst, err
	}
	return uint8(minorVersionUint), nil
}

func getTendermintPatchVersionFromSource() (uint8, error) {
	patchVersionUint, err := strconv.ParseUint(tendermint_version.Fix, 10, 8)
	if err != nil {
		return tendermintVersionPatchConst, err
	}
	return uint8(patchVersionUint), nil
}
