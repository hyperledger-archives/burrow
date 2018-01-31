package version

// Version represents the version of UUID. See page 7 in RFC 4122.
type Version byte

// Version List
const (
	Unknown       Version = iota // Unknwon
	TimeBased                    // V1: The time-based version
	DCESecurity                  // V2: The DCE security version, with embedded POSIX UIDs
	NameBasedMD5                 // V3: The name-based version that uses MD5 hashing
	Random                       // V4: The randomly or pseudo-randomly generated version
	NameBasedSHA1                // V5: The name-based version that uses SHA-1 hashing
)
