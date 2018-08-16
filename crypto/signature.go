package crypto

import (
	"fmt"

	"github.com/tmthrgd/go-hex"
	"golang.org/x/crypto/ed25519"
)

func SignatureFromBytes(bs []byte, curveType CurveType) (Signature, error) {
	switch curveType {
	case CurveTypeEd25519:
		if len(bs) != ed25519.SignatureSize {
			return Signature{}, fmt.Errorf("bytes passed have length %v by ed25519 signatures have %v bytes",
				len(bs), ed25519.SignatureSize)
		}
	case CurveTypeSecp256k1:
		// TODO: validate?
	}

	return Signature{CurveType: curveType, Signature: bs}, nil
}

func (sig Signature) RawBytes() []byte {
	return sig.Signature
}

func (sig Signature) String() string {
	return hex.EncodeUpperToString(sig.Signature)
}
