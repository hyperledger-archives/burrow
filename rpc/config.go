package rpc

type RPCConfig struct {
	V0 *TMConfig `json:",omitempty" toml:",omitempty"`
	TM *TMConfig `json:",omitempty" toml:",omitempty"`
}

type TMConfig struct {
	ListenAddress string
}

type V0Config struct {
}

func DefaultRPCConfig() *RPCConfig {
	return &RPCConfig{
		TM: DefaultTMConfig(),
	}
}

func DefaultTMConfig() *TMConfig {
	return &TMConfig{
		ListenAddress: ":46657",
	}
}
