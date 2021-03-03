package crypto

import (
	"fmt"
	"math/big"

	"github.com/btcsuite/btcd/btcec"
	hex "github.com/tmthrgd/go-hex"
	"golang.org/x/crypto/ed25519"
)

// Some big constnats
var big256 = big.NewInt(256)
var big2 = big.NewInt(2)

// https://github.com/ethereum/EIPs/blob/b3bbee93dc8a775af6a6b2525c9ac5f70a7e5710/EIPS/eip-155.md
var ethereumRecoveryIDOffset = big.NewInt(35)

// https://github.com/btcsuite/btcd/blob/7bbd9b0284de8492ae738ad8d722772925fa5a86/btcec/signature.go#L349
var btcecRecoveryIDOffset = big.NewInt(27)

type Secp256k1Signature struct {
	// Magic parity byte (value varies by implementation to carry additional information)
	V big.Int `json:"v"`
	R big.Int `json:"r"`
	S big.Int `json:"s"`
}

// Returns either 0 or 1 for the underlying parity of the public key solution
func (s *Secp256k1Signature) RecoveryIndex() uint {
	// odd  V => parity = 0
	// even V => parity = 1
	return s.V.Bit(0) ^ 0x01
}

func (s *Secp256k1Signature) BigRecoveryIndex() *big.Int {
	return new(big.Int).SetInt64(int64(s.RecoveryIndex()))
}

type EIP155Signature struct {
	Secp256k1Signature
}

// Get btcec compact signature (our standard)
func (s *EIP155Signature) ToCompactSignature() ([]byte, error) {
	v := new(big.Int)
	v.Add(s.BigRecoveryIndex(), btcecRecoveryIDOffset)
	compactSig := &CompactSecp256k1Signature{
		Secp256k1Signature{
			V: *v,
			R: s.R,
			S: s.S,
		},
	}
	bs, err := compactSig.Marshal()
	if err != nil {
		return nil, err
	}
	return bs, nil
}

/**
  btcec layout is:
  input:  [  v   |  r   |  s   ]
  bytes:  [  1   |  32  |  32  ]
  Where:
    v = 27 + recovery id (which of 4 possible x coords do we take as public key) (single byte but padded)
    r = encrypted random point
    s = signature proof

  Signature layout required by btcec:
  sig:    [  r   |  s   |  v  ]
  bytes:  [  32  |  32  |  1  ]
*/
type CompactSecp256k1Signature struct {
	Secp256k1Signature
}

func (s *CompactSecp256k1Signature) Marshal() ([]byte, error) {
	bs := make([]byte, btcec.PubKeyBytesLenUncompressed)
	bs[0] = byte(new(big.Int).Mod(&s.V, big256).Uint64())
	copy(bs[1:33], s.R.Bytes())
	copy(bs[33:], s.S.Bytes())
	return bs, nil
}

func (s *CompactSecp256k1Signature) Unmarshal(bs []byte) error {
	if len(bs) != btcec.PubKeyBytesLenUncompressed {
		return fmt.Errorf("must get uncompressed compact layout signature of %v bytes but got %v bytes",
			btcec.PubKeyBytesLenUncompressed, len(bs))
	}
	s.V.SetBytes(bs[0:1])
	s.R.SetBytes(bs[1:33])
	s.S.SetBytes(bs[33:])
	return nil
}

func SignatureFromBytes(bs []byte, curveType CurveType) (*Signature, error) {
	switch curveType {
	case CurveTypeEd25519:
		if len(bs) != ed25519.SignatureSize {
			return nil, fmt.Errorf("bytes passed have length %v by ed25519 signatures have %v bytes",
				len(bs), ed25519.SignatureSize)
		}
	case CurveTypeSecp256k1:
		// TODO: validate?
	}

	return &Signature{CurveType: curveType, Signature: bs}, nil
}

func (sig *Signature) RawBytes() []byte {
	return sig.Signature
}

func (sig *Signature) String() string {
	return hex.EncodeUpperToString(sig.Signature)
}

func GetEthChainID(chainID string) *big.Int {
	b := new(big.Int)
	id, ok := b.SetString(chainID, 10)
	if ok {
		return id
	}
	return b.SetBytes([]byte(chainID))
}

func GetEthSignatureRecoveryID(chainID string, parity *big.Int) *big.Int {
	// https://github.com/ethereum/EIPs/blob/b3bbee93dc8a775af6a6b2525c9ac5f70a7e5710/EIPS/eip-155.md
	v := new(big.Int)
	v.Mul(GetEthChainID(chainID), big2)
	v.Add(v, parity)
	v.Add(v, ethereumRecoveryIDOffset)
	return v
}

func (sig *Signature) GetEthSignature(chainID string) (*EIP155Signature, error) {
	if sig.CurveType != CurveTypeSecp256k1 {
		return nil, fmt.Errorf("can only GetEthSignature for %v keys, but got %v",
			CurveTypeSecp256k1, sig.CurveType)
	}
	compactSig := new(CompactSecp256k1Signature)
	err := compactSig.Unmarshal(sig.Signature)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal compact secp256k1 signature: %w", err)
	}
	return &EIP155Signature{
		Secp256k1Signature{
			V: *GetEthSignatureRecoveryID(chainID, compactSig.BigRecoveryIndex()),
			R: compactSig.R,
			S: compactSig.S,
		},
	}, nil
}
