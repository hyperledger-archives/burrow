package crypto

import (
	"bytes"
	"crypto/ecdsa"
	cryptoRand "crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/btcsuite/btcd/btcec"
	"golang.org/x/crypto/ed25519"
)

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
	case CurveTypeUnset:
		if len(bs) > 0 {
			return PublicKey{}, fmt.Errorf("attempting to create an 'unset' PublicKey but passed non-empty key bytes: %X", bs)
		}
		return PublicKey{}, nil
	default:
		return PublicKey{}, ErrInvalidCurve(curveType)
	}

	return PublicKey{PublicKey: bs, CurveType: curveType}, nil
}

func (p PrivateKey) RawBytes() []byte {
	return p.PrivateKey
}

func (p PrivateKey) Sign(msg []byte) (*Signature, error) {
	switch p.CurveType {
	case CurveTypeEd25519:
		if len(p.PrivateKey) != ed25519.PrivateKeySize {
			return nil, fmt.Errorf("bytes passed have length %v but ed25519 private keys have %v bytes",
				len(p.PrivateKey), ed25519.PrivateKeySize)
		}
		privKey := ed25519.PrivateKey(p.PrivateKey)
		return &Signature{CurveType: CurveTypeEd25519, Signature: ed25519.Sign(privKey, msg)}, nil
	case CurveTypeSecp256k1:
		if len(p.PrivateKey) != btcec.PrivKeyBytesLen {
			return nil, fmt.Errorf("bytes passed have length %v but secp256k1 private keys have %v bytes",
				len(p.PrivateKey), btcec.PrivKeyBytesLen)
		}
		privKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), p.PrivateKey)
		sig, err := privKey.Sign(Keccak256(msg))
		if err != nil {
			return nil, err
		}
		return &Signature{CurveType: CurveTypeSecp256k1, Signature: sig.Serialize()}, nil
	default:
		return nil, ErrInvalidCurve(p.CurveType)
	}
}

func (p PrivateKey) GetPublicKey() PublicKey {
	return PublicKey{CurveType: p.CurveType, PublicKey: p.PublicKey}
}

func (p PrivateKey) String() string {
	return fmt.Sprintf("PrivateKey<PublicKey:%X>", p.PublicKey)
}

func PrivateKeyFromRawBytes(privateKeyBytes []byte, curveType CurveType) (PrivateKey, error) {
	const ed25519PublicKeyOffset = ed25519.PrivateKeySize - ed25519.PublicKeySize
	switch curveType {
	case CurveTypeEd25519:
		if len(privateKeyBytes) != ed25519.PrivateKeySize {
			return PrivateKey{}, fmt.Errorf("bytes passed have length %v but ed25519 private keys have %v bytes",
				len(privateKeyBytes), ed25519.PrivateKeySize)
		}
		return PrivateKey{PrivateKey: privateKeyBytes, PublicKey: privateKeyBytes[ed25519PublicKeyOffset:], CurveType: CurveTypeEd25519}, nil
	case CurveTypeSecp256k1:
		if len(privateKeyBytes) != btcec.PrivKeyBytesLen {
			return PrivateKey{}, fmt.Errorf("bytes passed have length %v but secp256k1 private keys have %v bytes",
				len(privateKeyBytes), btcec.PrivKeyBytesLen)
		}
		_, publicKey := btcec.PrivKeyFromBytes(btcec.S256(), privateKeyBytes)
		return PrivateKey{PrivateKey: privateKeyBytes, PublicKey: publicKey.SerializeCompressed(), CurveType: CurveTypeSecp256k1}, nil
	default:
		return PrivateKey{}, ErrInvalidCurve(curveType)
	}
}

func GeneratePrivateKey(random io.Reader, curveType CurveType) (PrivateKey, error) {
	if random == nil {
		random = cryptoRand.Reader
	}
	switch curveType {
	case CurveTypeEd25519:
		_, privateKey, err := ed25519.GenerateKey(random)
		if err != nil {
			return PrivateKey{}, err
		}
		return PrivateKeyFromRawBytes(privateKey, CurveTypeEd25519)
	case CurveTypeSecp256k1:
		privateKey, err := ecdsa.GenerateKey(btcec.S256(), random)
		if err != nil {
			return PrivateKey{}, err
		}
		return PrivateKeyFromRawBytes(privateKey.D.Bytes(), CurveTypeSecp256k1)
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
