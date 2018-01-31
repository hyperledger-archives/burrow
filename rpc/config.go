package rpc

import "github.com/hyperledger/burrow/rpc/v0/server"

type RPCConfig struct {
	V0 *V0Config `json:",omitempty" toml:",omitempty"`
	TM *TMConfig `json:",omitempty" toml:",omitempty"`
}

type TMConfig struct {
	ListenAddress string
}

type V0Config struct {
	Server *server.ServerConfig
}

func DefaultRPCConfig() *RPCConfig {
	return &RPCConfig{
		TM: DefaultTMConfig(),
		V0: DefaultV0Config(),
	}
}
func DefaultV0Config() *V0Config {
	return &V0Config{
		Server: server.DefaultServerConfig(),
	}
}

func DefaultTMConfig() *TMConfig {
	return &TMConfig{
		ListenAddress: ":46657",
	}
}
