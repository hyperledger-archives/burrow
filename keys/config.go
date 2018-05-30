package keys

type KeysConfig struct {
	GRPCServiceEnabled bool
	RemoteAddress      string
	KeysDirectory      string
}

func DefaultKeysConfig() *KeysConfig {
	return &KeysConfig{
		// Default Monax keys port
		GRPCServiceEnabled: true,
		RemoteAddress:      "",
		KeysDirectory:      DefaultKeysDir,
	}
}
