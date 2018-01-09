package helpers

import (
	"bytes"

	"github.com/monax/keys/common"

	"github.com/tendermint/ed25519"
	"github.com/tendermint/go-wire"
	"golang.org/x/crypto/ripemd160"
)

type Signature interface {
	IsZero() bool
}

const (
	SignatureTypeEd25519 = byte(0x01)
)

type SignatureEd25519 [64]byte

func (sig SignatureEd25519) IsZero() bool { return len(sig) == 0 }

const (
	PrivKeyTypeEd25519 = byte(0x01)
)

type PrivKeyEd25519 [64]byte

const (
	PubKeyTypeEd25519 = byte(0x01)
)

func (key PrivKeyEd25519) Sign(msg []byte) Signature {
	privKeyBytes := [64]byte(key)
	signatureBytes := ed25519.Sign(&privKeyBytes, msg)
	return SignatureEd25519(*signatureBytes)
}

// Implements PubKey
type PubKeyEd25519 [32]byte

func (pubKey PubKeyEd25519) Address() []byte {
	w, n, err := new(bytes.Buffer), new(int), new(error)
	wire.WriteBinary(pubKey[:], w, n, err)
	if *err != nil {
		common.IfExit(*err)
	}
	// append type byte
	encodedPubkey := append([]byte{1}, w.Bytes()...)
	hasher := ripemd160.New()
	hasher.Write(encodedPubkey) // does not error
	return hasher.Sum(nil)
}

type (
	Hash [32]byte
)
