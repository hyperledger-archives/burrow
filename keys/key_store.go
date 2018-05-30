package keys

import (
	"github.com/hyperledger/burrow/crypto"
)

type Key struct {
	CurveType  crypto.CurveType
	Address    crypto.Address
	PublicKey  crypto.PublicKey
	PrivateKey crypto.PrivateKey
}

func NewKey(typ crypto.CurveType) (*Key, error) {
	privKey, err := crypto.GeneratePrivateKey(nil, typ)
	if err != nil {
		return nil, err
	}
	pubKey := privKey.GetPublicKey()
	return &Key{
		CurveType:  typ,
		PublicKey:  pubKey,
		Address:    pubKey.Address(),
		PrivateKey: privKey,
	}, nil
}

func (k *Key) Pubkey() []byte {
	return k.PublicKey.RawBytes()
}

func NewKeyFromPub(curveType crypto.CurveType, PubKeyBytes []byte) (*Key, error) {
	pubKey, err := crypto.PublicKeyFromBytes(PubKeyBytes, curveType)
	if err != nil {
		return nil, err
	}

	return &Key{
		CurveType: curveType,
		PublicKey: pubKey,
		Address:   pubKey.Address(),
	}, nil
}

func NewKeyFromPriv(curveType crypto.CurveType, PrivKeyBytes []byte) (*Key, error) {
	privKey, err := crypto.PrivateKeyFromRawBytes(PrivKeyBytes, curveType)

	if err != nil {
		return nil, err
	}

	pubKey := privKey.GetPublicKey()

	return &Key{
		CurveType:  curveType,
		Address:    pubKey.Address(),
		PublicKey:  pubKey,
		PrivateKey: privKey,
	}, nil
}

func (k *Key) Sign(hash []byte) ([]byte, error) {
	signature, err := k.PrivateKey.Sign(hash)
	if err != nil {
		return nil, err
	}
	return signature.RawBytes(), nil
}
