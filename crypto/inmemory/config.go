package inmemory

import (
	"fmt"
	acm "github.com/hyperledger/burrow/account"
)

type InMemoryCrypto struct {
	PrivateKey acm.PrivateKey
	PublicKey  acm.PublicKey
}

func DefaultInMemoryCryptoConfig() (*InMemoryCrypto, error) {
	privateKeyBytes, err := GenPrivateKey()

	if err != nil {
		return nil, fmt.Errorf("Unable to create default in memory config %v\n", err)
	}

	privateKey, err := acm.Ed25519PrivateKeyFromRawBytes(privateKeyBytes[:])
	if err != nil {
		return nil, fmt.Errorf("Unable to gen private key %v\n", err)
	}

	return &InMemoryCrypto{
		PrivateKey: privateKey,
		PublicKey:  privateKey.PublicKey(),
	}, nil
}
