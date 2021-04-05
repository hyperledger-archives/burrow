package rpc

import (
	"net"
)

// 'LocalHost' gets interpreted as ipv6
// TODO: revisit this
const LocalHost = "127.0.0.1"
const AnyLocal = "0.0.0.0"

type RPCConfig struct {
	Info     *ServerConfig  `json:",omitempty" toml:",omitempty"`
	Profiler *ServerConfig  `json:",omitempty" toml:",omitempty"`
	GRPC     *ServerConfig  `json:",omitempty" toml:",omitempty"`
	Metrics  *MetricsConfig `json:",omitempty" toml:",omitempty"`
	Web3     *ServerConfig  `json:",omitempty" toml:",omitempty"`
}

type ServerConfig struct {
	Enabled    bool
	ListenHost string
	ListenPort string
}

func (sc *ServerConfig) ListenAddress() string {
	return net.JoinHostPort(sc.ListenHost, sc.ListenPort)
}

type MetricsConfig struct {
	ServerConfig
	MetricsPath     string
	BlockSampleSize int
}

func DefaultRPCConfig() *RPCConfig {
	return &RPCConfig{
		Info:     DefaultInfoConfig(),
		Profiler: DefaultProfilerConfig(),
		GRPC:     DefaultGRPCConfig(),
		Metrics:  DefaultMetricsConfig(),
		Web3:     DefaultWeb3Config(),
	}
}

func DefaultInfoConfig() *ServerConfig {
	return &ServerConfig{
		Enabled:    true,
		ListenHost: AnyLocal,
		ListenPort: "26658",
	}
}

func DefaultGRPCConfig() *ServerConfig {
	return &ServerConfig{
		Enabled:    true,
		ListenHost: AnyLocal,
		ListenPort: "10997",
	}
}

func DefaultProfilerConfig() *ServerConfig {
	return &ServerConfig{
		Enabled:    false,
		ListenHost: AnyLocal,
		ListenPort: "6060",
	}
}

func DefaultMetricsConfig() *MetricsConfig {
	return &MetricsConfig{
		ServerConfig: ServerConfig{
			Enabled:    false,
			ListenHost: AnyLocal,
			ListenPort: "9102",
		},
		MetricsPath:     "/metrics",
		BlockSampleSize: 100,
	}
}

func DefaultWeb3Config() *ServerConfig {
	return &ServerConfig{
		Enabled:    true,
		ListenHost: AnyLocal,
		ListenPort: "26660",
	}
}
