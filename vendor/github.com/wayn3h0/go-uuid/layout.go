package uuid

import (
	"github.com/wayn3h0/go-uuid/internal/layout"
)

// Layout represents the layout of UUID. See page 5 in RFC 4122.
type Layout byte

// Layouts.
const (
	LayoutInvalid   = Layout(layout.Invalid)   // Invalid
	LayoutNCS       = Layout(layout.NCS)       // Reserved, NCS backward compatibility. (Values: 0x00-0x07)
	LayoutRFC4122   = Layout(layout.RFC4122)   // The variant specified in RFC 4122. (Values: 0x08-0x0b)
	LayoutMicrosoft = Layout(layout.Microsoft) // Reserved, Microsoft Corporation backward compatibility. (Values: 0x0c-0x0d)
	LayoutFuture    = Layout(layout.Future)    // Reserved for future definition. (Values: 0x0e-0x0f)
)

// String returns English description of layout.
func (this Layout) String() string {
	switch this {
	case LayoutNCS:
		return "Layout: Reserved For NCS"
	case LayoutRFC4122:
		return "Layout: RFC 4122"
	case LayoutMicrosoft:
		return "Layout: Reserved For Microsoft"
	case LayoutFuture:
		return "Layout: Reserved For Future"
	default:
		return "Layout: Invalid"
	}
}
