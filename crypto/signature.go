package crypto

import (
	"fmt"

	"golang.org/x/crypto/ed25519"
)

type Signature struct {
	Signature []byte
}

// Currently this is a stub that reads the raw bytes returned by key_client and returns
// an ed25519 signature.
func SignatureFromBytes(bs []byte, curveType CurveType) (Signature, error) {
	switch curveType {
	case CurveTypeEd25519:
		signatureEd25519 := Signature{}
		if len(bs) != ed25519.SignatureSize {
			return Signature{}, fmt.Errorf("bytes passed have length %v by ed25519 signatures have %v bytes",
				len(bs), ed25519.SignatureSize)
		}
		copy(signatureEd25519.Signature[:], bs)
		return Signature{
			Signature: bs,
		}, nil
	case CurveTypeSecp256k1:
		return Signature{
			Signature: bs,
		}, nil
	default:
		return Signature{}, nil
	}
}

func (sig Signature) RawBytes() []byte {
	return sig.Signature
}

func ChainSign(signer Signer, chainID string, o Signable) (Signature, error) {
	sig, err := signer.Sign(SignBytes(chainID, o))
	if err != nil {
		return Signature{}, err
	}
	return sig, nil
}
