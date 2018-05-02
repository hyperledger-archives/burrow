package crypto

import (
	"bytes"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/crypto/ripemd160"

	"github.com/btcsuite/btcd/btcec"
	tm_crypto "github.com/tendermint/go-crypto"
	"golang.org/x/crypto/ed25519"
)

type CurveType int8

const (
	CurveTypeSecp256k1 CurveType = iota
	CurveTypeEd25519
)

func (k CurveType) String() string {
	switch k {
	case CurveTypeSecp256k1:
		return "secp256k1"
	case CurveTypeEd25519:
		return "ed25519"
	default:
		return "unknown"
	}
}

func CurveTypeFromString(s string) (CurveType, error) {
	switch s {
	case "secp256k1":
		return CurveTypeSecp256k1, nil
	case "ed25519":
		return CurveTypeEd25519, nil
	default:
		var k CurveType
		return k, ErrInvalidCurve(s)
	}
}

type ErrInvalidCurve string

func (err ErrInvalidCurve) Error() string {
	return fmt.Sprintf("invalid curve type")
}

// The types in this file allow us to control serialisation of keys and signatures, as well as the interface
// exposed regardless of crypto library

type Signer interface {
	Sign(msg []byte) (Signature, error)
}

// PublicKey
type PublicKey struct {
	CurveType CurveType
	PublicKey []byte
}

type PrivateKey struct {
	CurveType  CurveType
	PublicKey  []byte
	PrivateKey []byte
}

type PublicKeyJSON struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

func (p PublicKey) MarshalJSON() ([]byte, error) {
	jStruct := PublicKeyJSON{
		Type: p.CurveType.String(),
		Data: fmt.Sprintf("%X", p.PublicKey),
	}
	txt, err := json.Marshal(jStruct)
	return txt, err
}

func (p PublicKey) MarshalText() ([]byte, error) {
	return p.MarshalJSON()
}

func (p *PublicKey) UnmarshalJSON(text []byte) error {
	var jStruct PublicKeyJSON
	err := json.Unmarshal(text, &jStruct)
	if err != nil {
		return err
	}
	CurveType, err := CurveTypeFromString(jStruct.Type)
	if err != nil {
		return err
	}
	bs, err := hex.DecodeString(jStruct.Data)
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

func (p PublicKey) IsValid() bool {
	switch p.CurveType {
	case CurveTypeEd25519:
		return len(p.PublicKey) == ed25519.PublicKeySize
	case CurveTypeSecp256k1:
		return len(p.PublicKey) == btcec.PubKeyBytesLenCompressed
	default:
		return false
	}
}
func (p PublicKey) Verify(msg []byte, signature Signature) bool {
	switch p.CurveType {
	case CurveTypeEd25519:
		return ed25519.Verify(p.PublicKey, msg, signature.Signature[:])
	case CurveTypeSecp256k1:
		pub, err := btcec.ParsePubKey(p.PublicKey, btcec.S256())
		if err != nil {
			return false
		}
		sig, err := btcec.ParseDERSignature(signature.Signature, btcec.S256())
		if err != nil {
			return false
		}
		return sig.Verify(msg, pub)
	default:
		return false
	}
}

func (p PublicKey) Address() Address {
	switch p.CurveType {
	case CurveTypeEd25519:
		// FIMXE: tendermint go-crypto-0.5.0 uses weird scheme, this is fixed in 0.6.0
		tmPubKey := new(tm_crypto.PubKeyEd25519)
		copy(tmPubKey[:], p.PublicKey)
		addr, _ := AddressFromBytes(tmPubKey.Address())
		return addr
	case CurveTypeSecp256k1:
		sha := sha256.New()
		sha.Write(p.PublicKey[:])

		hash := ripemd160.New()
		hash.Write(sha.Sum(nil))
		addr, _ := AddressFromBytes(hash.Sum(nil))
		return addr
	default:
		panic(fmt.Sprintf("unknown CurveType %d", p.CurveType))
	}
}

func (p PublicKey) RawBytes() []byte {
	return p.PublicKey[:]
}

func (p PublicKey) String() string {
	return fmt.Sprintf("%X", p.PublicKey[:])
}

// Signable is an interface for all signable things.
// It typically removes signatures before serializing.
type Signable interface {
	WriteSignBytes(chainID string, w io.Writer, n *int, err *error)
}

// SignBytes is a convenience method for getting the bytes to sign of a Signable.
func SignBytes(chainID string, o Signable) []byte {
	buf, n, err := new(bytes.Buffer), new(int), new(error)
	o.WriteSignBytes(chainID, buf, n, err)
	if *err != nil {
		panic(fmt.Sprintf("could not write sign bytes for a signable: %s", *err))
	}
	return buf.Bytes()
}

// Currently this is a stub that reads the raw bytes returned by key_client and returns
// an ed25519 public key.
func PublicKeyFromBytes(bs []byte, curveType CurveType) (PublicKey, error) {
	switch curveType {
	case CurveTypeEd25519:
		if len(bs) != ed25519.PublicKeySize {
			return PublicKey{}, fmt.Errorf("bytes passed have length %v but ed25519 public keys have %v bytes",
				len(bs), ed25519.PublicKeySize)
		}
	case CurveTypeSecp256k1:
		if len(bs) != btcec.PubKeyBytesLenCompressed {
			return PublicKey{}, fmt.Errorf("bytes passed have length %v but secp256k1 public keys have %v bytes",
				len(bs), btcec.PubKeyBytesLenCompressed)
		}
	default:
		return PublicKey{}, ErrInvalidCurve(curveType)
	}

	return PublicKey{PublicKey: bs, CurveType: curveType}, nil
}

func (p PrivateKey) RawBytes() []byte {
	return p.PrivateKey
}

func (p PrivateKey) Sign(msg []byte) (Signature, error) {
	switch p.CurveType {
	case CurveTypeEd25519:
		if len(p.PrivateKey) != ed25519.PrivateKeySize {
			return Signature{}, fmt.Errorf("bytes passed have length %v but ed25519 private keys have %v bytes",
				len(p.PrivateKey), ed25519.PrivateKeySize)
		}
		privKey := ed25519.PrivateKey(p.PrivateKey)
		return Signature{ed25519.Sign(privKey, msg)}, nil
	case CurveTypeSecp256k1:
		if len(p.PrivateKey) != btcec.PrivKeyBytesLen {
			return Signature{}, fmt.Errorf("bytes passed have length %v but secp256k1 private keys have %v bytes",
				len(p.PrivateKey), btcec.PrivKeyBytesLen)
		}
		privKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), p.PrivateKey)

		sig, err := privKey.Sign(msg)
		if err != nil {
			return Signature{}, err
		}
		return Signature{Signature: sig.Serialize()}, nil
	default:
		return Signature{}, ErrInvalidCurve(p.CurveType)
	}
}

func (p PrivateKey) GetPublicKey() PublicKey {
	return PublicKey{CurveType: p.CurveType, PublicKey: p.PublicKey}
}

func PrivateKeyFromRawBytes(privKeyBytes []byte, curveType CurveType) (PrivateKey, error) {
	switch curveType {
	case CurveTypeEd25519:
		if len(privKeyBytes) != ed25519.PrivateKeySize {
			return PrivateKey{}, fmt.Errorf("bytes passed have length %v but ed25519 private keys have %v bytes",
				len(privKeyBytes), ed25519.PrivateKeySize)
		}
		return PrivateKey{PrivateKey: privKeyBytes, PublicKey: privKeyBytes[32:], CurveType: CurveTypeEd25519}, nil
	case CurveTypeSecp256k1:
		if len(privKeyBytes) != btcec.PrivKeyBytesLen {
			return PrivateKey{}, fmt.Errorf("bytes passed have length %v but secp256k1 private keys have %v bytes",
				len(privKeyBytes), btcec.PrivKeyBytesLen)
		}
		privKey, pubKey := btcec.PrivKeyFromBytes(btcec.S256(), privKeyBytes)
		return PrivateKey{PrivateKey: privKey.Serialize(), PublicKey: pubKey.SerializeCompressed(), CurveType: CurveTypeSecp256k1}, nil
	default:
		return PrivateKey{}, ErrInvalidCurve(curveType)
	}
}

func GeneratePrivateKey(random io.Reader, curveType CurveType) (PrivateKey, error) {
	if random == nil {
		random = crand.Reader
	}
	switch curveType {
	case CurveTypeEd25519:
		_, priv, err := ed25519.GenerateKey(random)
		if err != nil {
			return PrivateKey{}, err
		}
		return PrivateKeyFromRawBytes(priv, CurveTypeEd25519)
	case CurveTypeSecp256k1:
		privKeyBytes := make([]byte, 32)
		_, err := random.Read(privKeyBytes)
		if err != nil {
			return PrivateKey{}, err
		}
		return PrivateKeyFromRawBytes(privKeyBytes, CurveTypeSecp256k1)
	default:
		return PrivateKey{}, ErrInvalidCurve(curveType)
	}
}

func PrivateKeyFromSecret(secret string, curveType CurveType) PrivateKey {
	hasher := sha256.New()
	hasher.Write(([]byte)(secret))
	// No error from a buffer
	privateKey, _ := GeneratePrivateKey(bytes.NewBuffer(hasher.Sum(nil)), curveType)
	return privateKey
}

// Ensures the last 32 bytes of the ed25519 private key is the public key derived from the first 32 private bytes
func EnsureEd25519PrivateKeyCorrect(candidatePrivateKey ed25519.PrivateKey) error {
	if len(candidatePrivateKey) != ed25519.PrivateKeySize {
		return fmt.Errorf("ed25519 key has size %v but %v bytes passed as key", ed25519.PrivateKeySize,
			len(candidatePrivateKey))
	}
	_, derivedPrivateKey, err := ed25519.GenerateKey(bytes.NewBuffer(candidatePrivateKey))
	if err != nil {
		return err
	}
	if !bytes.Equal(derivedPrivateKey, candidatePrivateKey) {
		return fmt.Errorf("ed25519 key generated from prefix of %X should equal %X, but is %X",
			candidatePrivateKey, candidatePrivateKey, derivedPrivateKey)
	}
	return nil
}

func ChainSign(signer Signer, chainID string, o Signable) (Signature, error) {
	sig, err := signer.Sign(SignBytes(chainID, o))
	if err != nil {
		return Signature{}, err
	}
	return sig, nil
}

// Signature

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
