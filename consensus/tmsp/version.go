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

package tmsp

import (
	version "github.com/eris-ltd/eris-db/version"
)

const (
	// Client identifier to advertise over the network
	tmspClientIdentifier = "tmsp"
	// Major version component of the current release
	tmspVersionMajor = 0
	// Minor version component of the current release
	tmspVersionMinor = 6
	// Patch version component of the current release
	tmspVersionPatch = 0
)

func GetTmspVersion() *version.VersionIdentifier {
	return version.New(tmspClientIdentifier, tmspVersionMajor, tmspVersionMinor,
		tmspVersionPatch)
}
