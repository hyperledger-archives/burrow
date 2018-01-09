package layout

// Layout represents the layout of UUID. See page 5 in RFC 4122.
type Layout byte

const (
	Invalid   Layout = iota // Invalid
	NCS                     // Reserved, NCS backward compatibility. (Values: 0x00-0x07)
	RFC4122                 // The variant specified in RFC 4122. (Values: 0x08-0x0b)
	Microsoft               // Reserved, Microsoft Corporation backward compatibility. (Values: 0x0c-0x0d)
	Future                  // Reserved for future definition. (Values: 0x0e-0x0f)
)
