package uuid

// Style represents the style of UUID string.
type Style byte

// Styles.
const (
	StyleStandard    Style = iota + 1 // xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx (8-4-4-4-12, length: 36)
	StyleWithoutDash                  // xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx (length: 32)
)

// String returns English description of style.
func (this Style) String() string {
	switch this {
	case StyleStandard:
		return "Style: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx (8-4-4-4-12)"
	case StyleWithoutDash:
		return "Style: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	default:
		return "Style: Unknown"
	}
}
