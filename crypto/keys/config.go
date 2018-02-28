package keys

import (
	acm "github.com/hyperledger/burrow/account"
)

type KeysConfig struct {
	URL              string       `json:",omitempty" toml:",omitempty"`
	ValidatorAddress *acm.Address `json:",omitempty" toml:",omitempty"`
}

func DefaultKeysConfig() *KeysConfig {
	return &KeysConfig{
		// Default Monax keys port
		URL: "http://localhost:4767",
	}
}
