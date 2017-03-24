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

package tendermint

import (
	"strconv"

	tendermint_version "github.com/tendermint/tendermint/version"

	version "github.com/monax/eris-db/version"
)

const (
	// Client identifier to advertise over the network
	tendermintClientIdentifier = "tendermint"
	// Major version component of the current release
	tendermintVersionMajorConst uint8 = 0
	// Minor version component of the current release
	tendermintVersionMinorConst uint8 = 8
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
