package keys

type KeysConfig struct {
	URL string
}

func DefaultKeysConfig() *KeysConfig {
	return &KeysConfig{
		// Default Monax keys port
		URL: "http://localhost:4767",
	}
}
