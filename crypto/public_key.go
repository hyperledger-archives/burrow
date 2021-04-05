package crypto

import (
	"encoding/json"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/tendermint/tendermint/crypto/tmhash"
	hex "github.com/tmthrgd/go-hex"
	"golang.org/x/crypto/ed25519"
)

type PublicKeyJSON struct {
	CurveType string
	PublicKey string
}

// Returns the length in bytes of the public key
func PublicKeyLength(curveType CurveType) int {
	switch curveType {
	case CurveTypeEd25519:
		return ed25519.PublicKeySize
	case CurveTypeSecp256k1:
		return btcec.PubKeyBytesLenUncompressed
	default:
		// Other functions rely on this
		return 0
	}
}

func (p *PublicKey) IsSet() bool {
	return p != nil && p.CurveType != CurveTypeUnset && p.IsValid()
}

func (p *PublicKey) MarshalJSON() ([]byte, error) {
	jStruct := PublicKeyJSON{
		CurveType: p.CurveType.String(),
		PublicKey: hex.EncodeUpperToString(p.PublicKey),
	}
	txt, err := json.Marshal(jStruct)
	return txt, err
}

func (p *PublicKey) MarshalText() ([]byte, error) {
	return p.MarshalJSON()
}

func (p *PublicKey) UnmarshalJSON(text []byte) error {
	var jStruct PublicKeyJSON
	err := json.Unmarshal(text, &jStruct)
	if err != nil {
		return err
	}
	CurveType, err := CurveTypeFromString(jStruct.CurveType)
	if err != nil {
		return err
	}
	bs, err := hex.DecodeString(jStruct.PublicKey)
	if err != nil {
		return err
	}
	p.CurveType = CurveType
	p.PublicKey = bs
	return nil
}

func (p *PublicKey) UnmarshalText(text []byte) error {
	return p.UnmarshalJSON(text)
}

func (p *PublicKey) IsValid() bool {
	publicKeyLength := PublicKeyLength(p.CurveType)
	return publicKeyLength != 0 && publicKeyLength == len(p.PublicKey)
}

func (p *PublicKey) Verify(msg []byte, signature *Signature) error {
	switch p.CurveType {
	case CurveTypeUnset:
		return fmt.Errorf("public key is unset")
	case CurveTypeEd25519:
		if ed25519.Verify(p.PublicKey.Bytes(), msg, signature.Signature) {
			return nil
		}
		return fmt.Errorf("signature '%X' is not a valid ed25519 signature for message: %s",
			signature.Signature, string(msg))
	case CurveTypeSecp256k1:
		_, _, err := btcec.RecoverCompact(btcec.S256(), signature.Signature, Keccak256(msg))
		if err != nil {
			return fmt.Errorf("signature verification for secp256k1 key failed: %v", err)
		}
		return nil
	default:
		return fmt.Errorf("invalid curve type")
	}
}

func (p *PublicKey) GetAddress() Address {
	switch p.CurveType {
	case CurveTypeEd25519:
		addr, _ := AddressFromBytes(tmhash.Sum(p.PublicKey))
		return addr
	case CurveTypeSecp256k1:
		pub, err := btcec.ParsePubKey(p.PublicKey.Bytes(), btcec.S256())
		if err != nil {
			return Address{}
		}
		hash := Keccak256(pub.SerializeUncompressed()[1:])
		addr, _ := AddressFromBytes(hash[len(hash)-AddressLength:])
		return addr
	default:
		panic(fmt.Sprintf("unknown CurveType %d", p.CurveType))
	}
}

func (p *PublicKey) AddressHashType() string {
	switch p.CurveType {
	case CurveTypeEd25519:
		return "go-crypto-0.5.0"
	case CurveTypeSecp256k1:
		return "btc"
	default:
		return ""
	}
}

func (p *PublicKey) String() string {
	return hex.EncodeUpperToString(p.PublicKey)
}
