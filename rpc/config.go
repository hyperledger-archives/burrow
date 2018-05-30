package rpc

import "github.com/hyperledger/burrow/rpc/v0/server"

type RPCConfig struct {
	V0       *V0Config       `json:",omitempty" toml:",omitempty"`
	TM       *TMConfig       `json:",omitempty" toml:",omitempty"`
	Profiler *ProfilerConfig `json:",omitempty" toml:",omitempty"`
	GRPC     *GRPCConfig     `json:",omitempty" toml:",omitempty"`
}

type TMConfig struct {
	Enabled       bool
	ListenAddress string
}

type V0Config struct {
	Enabled bool
	Server  *server.ServerConfig
}

type ProfilerConfig struct {
	Enabled       bool
	ListenAddress string
}

type GRPCConfig struct {
	Enabled       bool
	ListenAddress string
}

func DefaultRPCConfig() *RPCConfig {
	return &RPCConfig{
		TM:       DefaultTMConfig(),
		V0:       DefaultV0Config(),
		Profiler: DefaultProfilerConfig(),
		GRPC:     DefaultGRPCConfig(),
	}
}

func DefaultV0Config() *V0Config {
	return &V0Config{
		Enabled: true,
		Server:  server.DefaultServerConfig(),
	}
}

func DefaultTMConfig() *TMConfig {
	return &TMConfig{
		Enabled:       true,
		ListenAddress: "tcp://localhost:46657",
	}
}

func DefaultGRPCConfig() *GRPCConfig {
	return &GRPCConfig{
		Enabled:       true,
		ListenAddress: "localhost:10997",
	}
}

func DefaultProfilerConfig() *ProfilerConfig {
	return &ProfilerConfig{
		Enabled:       false,
		ListenAddress: "tcp://localhost:6060",
	}
}
