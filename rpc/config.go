package rpc

import "fmt"

// 'localhost' gets interpreted as ipv6
// TODO: revisit this
const localhost = "127.0.0.1"

type RPCConfig struct {
	Info     *ServerConfig  `json:",omitempty" toml:",omitempty"`
	Profiler *ServerConfig  `json:",omitempty" toml:",omitempty"`
	GRPC     *ServerConfig  `json:",omitempty" toml:",omitempty"`
	Metrics  *MetricsConfig `json:",omitempty" toml:",omitempty"`
}

type ServerConfig struct {
	Enabled       bool
	ListenAddress string
}

type ProfilerConfig struct {
	Enabled       bool
	ListenAddress string
}

type GRPCConfig struct {
	Enabled       bool
	ListenAddress string
}

type MetricsConfig struct {
	Enabled         bool
	ListenAddress   string
	MetricsPath     string
	BlockSampleSize uint64
}

func DefaultRPCConfig() *RPCConfig {
	return &RPCConfig{
		Info:     DefaultInfoConfig(),
		Profiler: DefaultProfilerConfig(),
		GRPC:     DefaultGRPCConfig(),
		Metrics:  DefaultMetricsConfig(),
	}
}

func DefaultInfoConfig() *ServerConfig {
	return &ServerConfig{
		Enabled:       true,
		ListenAddress: fmt.Sprintf("tcp://%s:26658", localhost),
	}
}

func DefaultGRPCConfig() *ServerConfig {
	return &ServerConfig{
		Enabled:       true,
		ListenAddress: fmt.Sprintf("%s:10997", localhost),
	}
}

func DefaultProfilerConfig() *ServerConfig {
	return &ServerConfig{
		Enabled:       false,
		ListenAddress: fmt.Sprintf("tcp://%s:6060", localhost),
	}
}

func DefaultMetricsConfig() *MetricsConfig {
	return &MetricsConfig{
		Enabled:         false,
		ListenAddress:   fmt.Sprintf("tcp://%s:9102", localhost),
		MetricsPath:     "/metrics",
		BlockSampleSize: 100,
	}
}
