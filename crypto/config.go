package crypto

import (
	"github.com/hyperledger/burrow/crypto/inmemory"
	"github.com/hyperledger/burrow/crypto/keys"
)

type CryptoConfig struct {
	KeysServer     *keys.KeysConfig         `json:",omitempty" toml:",omitempty"`
	InMemoryCrypto *inmemory.InMemoryCrypto `json:",omitempty" toml:",omitempty"`
}

func DefaultCryptoConfig() *CryptoConfig {
	return &CryptoConfig{
		KeysServer:     keys.DefaultKeysConfig(),
		InMemoryCrypto: nil, // Default to keys server
	}
}
