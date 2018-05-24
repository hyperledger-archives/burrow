package mock

import (
	"crypto/rand"
	"encoding/json"
	"fmt"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/keys"
	"github.com/pkg/errors"
	"github.com/tendermint/go-crypto"
	"github.com/wayn3h0/go-uuid"
	"golang.org/x/crypto/ed25519"
)

// Mock ed25510 key for mock keys client
// Simple ed25519 key structure for mock purposes with ripemd160 address
type Key struct {
	Name       string
	Address    acm.Address
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

	var ed25519 crypto.PubKeyEd25519
	copy(ed25519[:], publicKey[:])

	key.Address, err = acm.AddressFromBytes(ed25519.Address())
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

func mockKeyFromPrivateAccount(privateAccount acm.PrivateAccount) *Key {
	_, ok := privateAccount.PrivateKey().Unwrap().(crypto.PrivKeyEd25519)
	if !ok {
		panic(fmt.Errorf("mock key client only supports ed25519 private keys at present"))
	}
	key := &Key{
		Name:       privateAccount.Address().String(),
		Address:    privateAccount.Address(),
		PublicKey:  privateAccount.PublicKey().RawBytes(),
		PrivateKey: privateAccount.PrivateKey().RawBytes(),
	}
	return key
}

func (key *Key) Sign(message []byte) (acm.Signature, error) {
	return acm.SignatureFromBytes(ed25519.Sign(key.PrivateKey, message))
}

// TODO: remove after merging keys taken from there to match serialisation
type plainKeyJSON struct {
	Id         []byte
	Type       string
	Address    string
	PrivateKey []byte
}

// Returns JSON string compatible with that stored by monax-keys
func (key *Key) MonaxKeysJSON() string {
	id, err := uuid.NewRandom()
	if err != nil {
		return errors.Wrap(err, "could not create monax key json").Error()
	}
	jsonKey := plainKeyJSON{
		Id:         []byte(id.String()),
		Address:    key.Address.String(),
		Type:       string(keys.KeyTypeEd25519Ripemd160),
		PrivateKey: key.PrivateKey,
	}
	bs, err := json.Marshal(jsonKey)
	if err != nil {
		return errors.Wrap(err, "could not create monax key json").Error()
	}
	return string(bs)
}
