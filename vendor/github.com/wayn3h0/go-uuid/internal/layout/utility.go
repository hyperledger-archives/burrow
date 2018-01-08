package layout

// Set sets the layout for uuid.
func Set(uuid []byte, layout Layout) {
	switch layout {
	case NCS:
		uuid[8] = (uuid[8] | 0x00) & 0x0f // Msb0=0
	case RFC4122:
		uuid[8] = (uuid[8] | 0x80) & 0x8f // Msb0=1, Msb1=0
	case Microsoft:
		uuid[8] = (uuid[8] | 0xc0) & 0xcf // Msb0=1, Msb1=1, Msb2=0
	case Future:
		uuid[8] = (uuid[8] | 0xe0) & 0xef // Msb0=1, Msb1=1, Msb2=1
	default:
		panic("uuid: layout is invalid")
	}
}

// Get returns layout of uuid.
func Get(uuid []byte) Layout {
	switch {
	case (uuid[8] & 0x80) == 0x00:
		return NCS
	case (uuid[8] & 0xc0) == 0x80:
		return RFC4122
	case (uuid[8] & 0xe0) == 0xc0:
		return Microsoft
	case (uuid[8] & 0xe0) == 0xe0:
		return Future
	}

	return Invalid
}
