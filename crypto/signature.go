package crypto

import (
	"fmt"

	"github.com/tmthrgd/go-hex"
	"golang.org/x/crypto/ed25519"
)

type Signature []byte

func SignatureFromBytes(bs []byte, curveType CurveType) (Signature, error) {
	switch curveType {
	case CurveTypeEd25519:
		var signatureEd25519 Signature
		if len(bs) != ed25519.SignatureSize {
			return nil, fmt.Errorf("bytes passed have length %v by ed25519 signatures have %v bytes",
				len(bs), ed25519.SignatureSize)
		}
		copy(signatureEd25519, bs)
		return bs, nil
	case CurveTypeSecp256k1:
		return bs, nil
	default:
		return nil, nil
	}
}

func (sig Signature) RawBytes() []byte {
	return sig
}

func (sig *Signature) UnmarshalText(hexBytes []byte) error {
	bs, err := hex.DecodeString(string(hexBytes))
	if err != nil {
		return err
	}
	*sig = bs
	return nil
}

func (sig Signature) MarshalText() ([]byte, error) {
	return []byte(sig.String()), nil
}

func (sig Signature) String() string {
	return hex.EncodeUpperToString(sig)
}

// Protobuf support
func (sig Signature) Marshal() ([]byte, error) {
	return sig, nil
}

func (sig *Signature) Unmarshal(data []byte) error {
	*sig = data
	return nil
}

func (sig Signature) MarshalTo(data []byte) (int, error) {
	return copy(data, sig), nil
}

func (sig Signature) Size() int {
	return len(sig)
}
