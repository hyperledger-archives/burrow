package version

// Set sets the version for uuid.
func Set(uuid []byte, version Version) {
	switch version {
	case TimeBased:
		uuid[6] = (uuid[6] | 0x10) & 0x1f
	case DCESecurity:
		uuid[6] = (uuid[6] | 0x20) & 0x2f
	case NameBasedMD5:
		uuid[6] = (uuid[6] | 0x30) & 0x3f
	case Random:
		uuid[6] = (uuid[6] | 0x40) & 0x4f
	case NameBasedSHA1:
		uuid[6] = (uuid[6] | 0x50) & 0x5f
	default:
		panic("uuid: version is unknown")
	}
}

// Get gets the version of uuid.
func Get(uuid []byte) Version {
	version := uuid[6] >> 4
	if version > 0 && version < 6 {
		return Version(version)
	}

	return Unknown
}
