package keys

type KeysConfig struct {
	ServerEnabled bool
	URL           string
}

func DefaultKeysConfig() *KeysConfig {
	return &KeysConfig{
		// Default Monax keys port
		ServerEnabled: true,
		URL:           "",
	}
}
