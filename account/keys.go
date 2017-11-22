package account

import "github.com/tendermint/go-crypto"

// This allows us to control serialisation

type PublicKey struct {
	crypto.PubKey
}

func PublicKeyFromPubKey(pubKey crypto.PubKey) PublicKey {
	return PublicKey{PubKey: pubKey}
}

func PrivateKeyFromPrivKey(privKey crypto.PrivKey) PrivateKey {
	return PrivateKey{PrivKey: privKey}
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

type PrivateKey struct {
	crypto.PrivKey
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
