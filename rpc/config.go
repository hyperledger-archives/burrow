package rpc

import "github.com/hyperledger/burrow/rpc/v0/server"

type RPCConfig struct {
	V0       *V0Config       `json:",omitempty" toml:",omitempty"`
	TM       *TMConfig       `json:",omitempty" toml:",omitempty"`
	Profiler *ProfilerConfig `json:",omitempty" toml:",omitempty"`
}

type TMConfig struct {
	Disabled      bool
	ListenAddress string
}

type V0Config struct {
	Disabled bool
	Server   *server.ServerConfig
}

type ProfilerConfig struct {
	Disabled      bool
	ListenAddress string
}

func DefaultRPCConfig() *RPCConfig {
	return &RPCConfig{
		TM:       DefaultTMConfig(),
		V0:       DefaultV0Config(),
		Profiler: DefaultProfilerConfig(),
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

func DefaultProfilerConfig() *ProfilerConfig {
	return &ProfilerConfig{
		Disabled:      true,
		ListenAddress: ":6060",
	}
}
