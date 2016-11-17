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

package version

import (
	"fmt"
)

// IMPORTANT: this version number needs to be manually kept
// in sync at the bottom of this file for the deployment scripts to parse
// the version number.
const (
	// Client identifier to advertise over the network
	erisClientIdentifier = "eris-db"
	// Major version component of the current release
	erisVersionMajor = 0
	// Minor version component of the current release
	erisVersionMinor = 12
	// Patch version component of the current release
	erisVersionPatch = 100
)

var erisVersion *VersionIdentifier

func init() {
	erisVersion = New(erisClientIdentifier, erisVersionMajor,
		erisVersionMinor, erisVersionPatch)
}

//------------------------------------------------------------------------------
// versioning globally for Eris-DB and scoped for modules

type VersionIdentifier struct {
	clientIdentifier string
	versionMajor     uint8
	versionMinor     uint8
	versionPatch     uint8
}

func New(client string, major, minor, patch uint8) *VersionIdentifier {
	v := new(VersionIdentifier)
	v.clientIdentifier = client
	v.versionMajor = major
	v.versionMinor = minor
	v.versionPatch = patch
	return v
}

// GetVersionString returns `client-major.minor.patch` for Eris-DB
// without a receiver, or for the version called on.
// MakeVersionString builds the same version string with provided parameters.
func GetVersionString() string { return erisVersion.GetVersionString() }
func (v *VersionIdentifier) GetVersionString() string {
	return fmt.Sprintf("%s-%d.%d.%d", v.clientIdentifier, v.versionMajor,
		v.versionMinor, v.versionPatch)
}

// note: the arguments are passed in as int (rather than uint8)
// because on asserting the version constructed from the configuration file
// the casting of an int to uint8 is uglier than expanding the type range here.
// Should the configuration file have an invalid integer (that could not convert)
// then this will equally be reflected in a failed assertion of the version string.
func MakeVersionString(client string, major, minor, patch int) string {
	return fmt.Sprintf("%s-%d.%d.%d", client, major, minor, patch)
}

// GetMinorVersionString returns `client-major.minor` for Eris-DB
// without a receiver, or for the version called on.
// MakeMinorVersionString builds the same version string with
// provided parameters.
func GetMinorVersionString() string { return erisVersion.GetVersionString() }
func (v *VersionIdentifier) GetMinorVersionString() string {
	return fmt.Sprintf("%s-%d.%d", v.clientIdentifier, v.versionMajor,
		v.versionMinor)
}

// note: similar remark applies here on the use of `int` over `uint8`
// for the arguments as above for MakeVersionString()
func MakeMinorVersionString(client string, major, minor, patch int) string {
	return fmt.Sprintf("%s-%d.%d", client, major, minor)
}

// GetVersion returns a tuple of client, major, minor, and patch as types,
// either for Eris-DB without a receiver or the called version structure.
func GetVersion() (client string, major, minor, patch uint8) {
	return erisVersion.GetVersion()
}
func (version *VersionIdentifier) GetVersion() (
	client string, major, minor, patch uint8) {
	return version.clientIdentifier, version.versionMajor, version.versionMinor,
		version.versionPatch
}

//------------------------------------------------------------------------------
// Matching functions

// MatchesMinorVersion matches the client identifier, major and minor version
// number of the reference version identifier to be equal with the receivers.
func MatchesMinorVersion(referenceVersion *VersionIdentifier) bool {
	return erisVersion.MatchesMinorVersion(referenceVersion)
}
func (version *VersionIdentifier) MatchesMinorVersion(
	referenceVersion *VersionIdentifier) bool {
	referenceClient, referenceMajor, referenceMinor, _ := referenceVersion.GetVersion()
	return version.clientIdentifier == referenceClient &&
		version.versionMajor == referenceMajor &&
		version.versionMinor == referenceMinor
}

//------------------------------------------------------------------------------
// Version number for tests/build_tool.sh

// IMPORTANT: Eris-DB version must be on the last line of this file for
// the deployment script tests/build_tool.sh to pick up the right label.
const VERSION = "0.12.100"
