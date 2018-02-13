package account

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/tendermint/go-crypto"
	"golang.org/x/crypto/ed25519"
)

// The types in this file allow us to control serialisation of keys and signatures, as well as the interface
// exposed regardless of crypto library

type Signer interface {
	Sign(msg []byte) (Signature, error)
}

// PublicKey
type PublicKey struct {
	crypto.PubKey `json:"unwrap"`
}

func PublicKeyFromGoCryptoPubKey(pubKey crypto.PubKey) (PublicKey, error) {
	_, err := AddressFromBytes(pubKey.Address())
	if err != nil {
		return PublicKey{}, fmt.Errorf("could not make valid address from public key %v: %v", pubKey, err)
	}
	return PublicKey{
		PubKey: pubKey,
	}, nil
}

// Currently this is a stub that reads the raw bytes returned by key_client and returns
// an ed25519 public key.
func PublicKeyFromBytes(bs []byte) (PublicKey, error) {
	//TODO: read a typed representation (most likely JSON) and do the right thing here
	// Only supports ed25519 currently so no switch on signature scheme
	pubKeyEd25519 := crypto.PubKeyEd25519{}
	if len(bs) != len(pubKeyEd25519) {
		return PublicKey{}, fmt.Errorf("bytes passed have length %v by ed25519 public keys have %v bytes",
			len(bs), len(pubKeyEd25519))
	}
	copy(pubKeyEd25519[:], bs)
	return PublicKeyFromGoCryptoPubKey(pubKeyEd25519.Wrap())
}

// Returns a copy of the raw untyped public key bytes
func (pk PublicKey) RawBytes() []byte {
	switch pubKey := pk.PubKey.Unwrap().(type) {
	case crypto.PubKeyEd25519:
		pubKeyCopy := crypto.PubKeyEd25519{}
		copy(pubKeyCopy[:], pubKey[:])
		return pubKeyCopy[:]
	case crypto.PubKeySecp256k1:
		pubKeyCopy := crypto.PubKeySecp256k1{}
		copy(pubKeyCopy[:], pubKey[:])
		return pubKeyCopy[:]
	default:
		return nil
	}
}

func (pk PublicKey) VerifyBytes(msg []byte, signature Signature) bool {
	return pk.PubKey.VerifyBytes(msg, signature.Signature)
}

func (pk PublicKey) Address() Address {
	// We check this on initialisation to avoid this panic, but returning an error here is ugly and caching
	// the address on PublicKey initialisation breaks go-wire serialisation since with unwrap we can only have one field.
	// We can do something better with better serialisation
	return MustAddressFromBytes(pk.PubKey.Address())
}

func (pk PublicKey) MarshalJSON() ([]byte, error) {
	return pk.PubKey.MarshalJSON()
}

func (pk *PublicKey) UnmarshalJSON(data []byte) error {
	return pk.PubKey.UnmarshalJSON(data)
}

func (pk PublicKey) MarshalText() ([]byte, error) {
	return pk.MarshalJSON()
}

func (pk *PublicKey) UnmarshalText(text []byte) error {
	return pk.UnmarshalJSON(text)
}

// PrivateKey

type PrivateKey struct {
	crypto.PrivKey `json:"unwrap"`
}

func PrivateKeyFromGoCryptoPrivKey(privKey crypto.PrivKey) (PrivateKey, error) {
	_, err := PublicKeyFromGoCryptoPubKey(privKey.PubKey())
	if err != nil {
		return PrivateKey{}, fmt.Errorf("could not create public key from private key: %v", err)
	}
	return PrivateKey{
		PrivKey: privKey,
	}, nil
}

func PrivateKeyFromSecret(secret string) PrivateKey {
	hasher := sha256.New()
	hasher.Write(([]byte)(secret))
	// No error from a buffer
	privateKey, _ := GeneratePrivateKey(bytes.NewBuffer(hasher.Sum(nil)))
	return privateKey
}

// Generates private key from a source of random bytes, if randomReader is nil crypto/rand.Reader is useds
func GeneratePrivateKey(randomReader io.Reader) (PrivateKey, error) {
	if randomReader == nil {
		randomReader = rand.Reader
	}
	_, ed25519PrivateKey, err := ed25519.GenerateKey(randomReader)
	if err != nil {
		return PrivateKey{}, err
	}
	return Ed25519PrivateKeyFromRawBytes(ed25519PrivateKey)
}

// Creates an ed25519 key from the raw private key bytes
func Ed25519PrivateKeyFromRawBytes(privKeyBytes []byte) (PrivateKey, error) {
	privKeyEd25519 := crypto.PrivKeyEd25519{}
	if len(privKeyBytes) != len(privKeyEd25519) {
		return PrivateKey{}, fmt.Errorf("bytes passed have length %v by ed25519 private keys have %v bytes",
			len(privKeyBytes), len(privKeyEd25519))
	}
	err := EnsureEd25519PrivateKeyCorrect(privKeyBytes)
	if err != nil {
		return PrivateKey{}, err
	}
	copy(privKeyEd25519[:], privKeyBytes)
	return PrivateKeyFromGoCryptoPrivKey(privKeyEd25519.Wrap())
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

func (pk PrivateKey) PublicKey() PublicKey {
	publicKey, err := PublicKeyFromGoCryptoPubKey(pk.PrivKey.PubKey())
	if err != nil {
		// We check this on initialisation to avoid this panic, but returning an error here is ugly and  caching
		// the public key on PrivateKey on initialisation breaks go-wire. We can do something better with better serialisation
		panic(fmt.Errorf("error making public key from private key: %v", publicKey))
	}
	return publicKey
}

// Returns a copy of the raw untyped private key bytes
func (pk PrivateKey) RawBytes() []byte {
	switch privKey := pk.PrivKey.Unwrap().(type) {
	case crypto.PrivKeyEd25519:
		privKeyCopy := crypto.PrivKeyEd25519{}
		copy(privKeyCopy[:], privKey[:])
		return privKeyCopy[:]
	case crypto.PrivKeySecp256k1:
		privKeyCopy := crypto.PrivKeySecp256k1{}
		copy(privKeyCopy[:], privKey[:])
		return privKeyCopy[:]
	default:
		return nil
	}
}

func (pk PrivateKey) Sign(msg []byte) (Signature, error) {
	return Signature{Signature: pk.PrivKey.Sign(msg)}, nil
}

func (pk PrivateKey) MarshalJSON() ([]byte, error) {
	return pk.PrivKey.MarshalJSON()
}

func (pk *PrivateKey) UnmarshalJSON(data []byte) error {
	return pk.PrivKey.UnmarshalJSON(data)
}

func (pk PrivateKey) MarshalText() ([]byte, error) {
	return pk.MarshalJSON()
}

func (pk *PrivateKey) UnmarshalText(text []byte) error {
	return pk.UnmarshalJSON(text)
}

// Signature

type Signature struct {
	crypto.Signature `json:"unwrap"`
}

func SignatureFromGoCryptoSignature(signature crypto.Signature) Signature {
	return Signature{Signature: signature}
}

// Currently this is a stub that reads the raw bytes returned by key_client and returns
// an ed25519 signature.
func SignatureFromBytes(bs []byte) (Signature, error) {
	//TODO: read a typed representation (most likely JSON) and do the right thing here
	// Only supports ed25519 currently so no switch on signature scheme
	signatureEd25519 := crypto.SignatureEd25519{}
	if len(bs) != len(signatureEd25519) {
		return Signature{}, fmt.Errorf("bytes passed have length %v by ed25519 signatures have %v bytes",
			len(bs), len(signatureEd25519))
	}
	copy(signatureEd25519[:], bs)
	return Signature{
		Signature: signatureEd25519.Wrap(),
	}, nil
}

func (s Signature) GoCryptoSignature() crypto.Signature {
	return s.Signature
}

func (s Signature) MarshalJSON() ([]byte, error) {
	return s.Signature.MarshalJSON()
}

func (s *Signature) UnmarshalJSON(data []byte) error {
	return s.Signature.UnmarshalJSON(data)
}

func (s Signature) MarshalText() ([]byte, error) {
	return s.MarshalJSON()
}

func (s *Signature) UnmarshalText(text []byte) error {
	return s.UnmarshalJSON(text)
}
