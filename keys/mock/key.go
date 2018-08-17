package mock

import (
	"crypto/rand"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/crypto"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ed25519"
)

// Mock ed25510 key for mock keys client
// Simple ed25519 key structure for mock purposes with ripemd160 address
type Key struct {
	Name       string
	Address    crypto.Address
	PublicKey  []byte
	PrivateKey []byte
}

func newKey(name string) (*Key, error) {
	key := &Key{
		Name:       name,
		PublicKey:  make([]byte, ed25519.PublicKeySize),
		PrivateKey: make([]byte, ed25519.PrivateKeySize),
	}
	// this is a mock key, so the entropy of the source is purely
	// for testing
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	copy(key.PrivateKey[:], privateKey[:])
	copy(key.PublicKey[:], publicKey[:])

	pk, err := crypto.PublicKeyFromBytes(publicKey, crypto.CurveTypeEd25519)
	if err != nil {
		return nil, err
	}

	key.Address, err = crypto.AddressFromBytes(pk.Address().Bytes())
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	if key.Name == "" {
		key.Name = key.Address.String()
	}
	return key, nil
}

func mockKeyFromPrivateAccount(privateAccount *acm.PrivateAccount) *Key {
	if privateAccount.PrivateKey().CurveType != crypto.CurveTypeEd25519 {
		panic(fmt.Errorf("mock key client only supports ed25519 private keys at present"))
	}
	key := &Key{
		Name:       privateAccount.Address().String(),
		Address:    privateAccount.Address(),
		PublicKey:  privateAccount.PublicKey().PublicKey,
		PrivateKey: privateAccount.PrivateKey().PrivateKey,
	}
	return key
}

func (key *Key) Sign(message []byte) (crypto.Signature, error) {
	return crypto.SignatureFromBytes(ed25519.Sign(key.PrivateKey, message), crypto.CurveTypeEd25519)
}

type PrivateKeyplainKeyJSON struct {
	Plain []byte
}

// TODO: remove after merging keys taken from there to match serialisation
type plainKeyJSON struct {
	Type       string
	Address    string
	PrivateKey PrivateKeyplainKeyJSON
}

// Returns JSON string compatible with that stored by monax-keys
func (key *Key) MonaxKeysJSON() string {
	jsonKey := plainKeyJSON{
		Address:    key.Address.String(),
		Type:       "ed25519",
		PrivateKey: PrivateKeyplainKeyJSON{Plain: key.PrivateKey},
	}
	bs, err := json.Marshal(jsonKey)
	if err != nil {
		return errors.Wrap(err, "could not create monax key json").Error()
	}
	return string(bs)
}
