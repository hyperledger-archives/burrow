package binary

import "github.com/tmthrgd/go-hex"

type HexBytes []byte

func (bs *HexBytes) UnmarshalText(hexBytes []byte) error {
	bs2, err := hex.DecodeString(string(hexBytes))
	if err != nil {
		return err
	}
	*bs = bs2
	return nil
}

func (bs HexBytes) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeUpperToString(bs)), nil
}
