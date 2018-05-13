package uuid

import (
	"github.com/wayn3h0/go-uuid/internal/version"
)

// Version represents the version of UUID. See page 7 in RFC 4122.
type Version byte

// Versions.
const (
	VersionUnknown       = Version(version.Unknown)       // Unknown
	VersionTimeBased     = Version(version.TimeBased)     // V1: The time-based version
	VersionDCESecurity   = Version(version.DCESecurity)   // V2: The DCE security version, with embedded POSIX UIDs
	VersionNameBasedMD5  = Version(version.NameBasedMD5)  // V3: The name-based version that uses MD5 hashing
	VersionRandom        = Version(version.Random)        // V4: The randomly or pseudo-randomly generated version
	VersionNameBasedSHA1 = Version(version.NameBasedSHA1) // V5: The name-based version that uses SHA-1 hashing
)

// Short names.
const (
	V1 = VersionTimeBased
	V2 = VersionDCESecurity
	V3 = VersionNameBasedMD5
	V4 = VersionRandom
	V5 = VersionNameBasedSHA1
)

// String returns English description of version.
func (this Version) String() string {
	switch this {
	case VersionTimeBased:
		return "Version 1: Time-Based"
	case VersionDCESecurity:
		return "Version 2: DCE Security With Embedded POSIX UIDs"
	case VersionNameBasedMD5:
		return "Version 3: Name-Based (MD5)"
	case VersionRandom:
		return "Version 4: Randomly OR Pseudo-Randomly Generated"
	case VersionNameBasedSHA1:
		return "Version 5: Name-Based (SHA-1)"
	default:
		return "Version: Unknown"
	}
}
