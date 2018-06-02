package rpc

import "github.com/hyperledger/burrow/rpc/v0/server"

type RPCConfig struct {
	V0       *V0Config     `json:",omitempty" toml:",omitempty"`
	TM       *ServerConfig `json:",omitempty" toml:",omitempty"`
	Profiler *ServerConfig `json:",omitempty" toml:",omitempty"`
	GRPC     *ServerConfig `json:",omitempty" toml:",omitempty"`
}

type ServerConfig struct {
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

func DefaultTMConfig() *ServerConfig {
	return &ServerConfig{
		Enabled:       true,
		ListenAddress: "tcp://localhost:46657",
	}
}

func DefaultGRPCConfig() *ServerConfig {
	return &ServerConfig{
		Enabled:       true,
		ListenAddress: "localhost:10997",
	}
}

func DefaultProfilerConfig() *ServerConfig {
	return &ServerConfig{
		Enabled:       false,
		ListenAddress: "tcp://localhost:6060",
	}
}
