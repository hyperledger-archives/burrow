package word

const Word160Length = 20

var Zero160 = Word160{}

type Word160 [Word160Length]byte

// Pad a Word160 on the left and embed it in a Word256 (as it is for account addresses in EVM)
func (w Word160) Word256() Word256 {
	return LeftPadWord256(w[:])
}
