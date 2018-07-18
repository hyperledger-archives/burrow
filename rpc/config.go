package rpc

type RPCConfig struct {
	TM       *ServerConfig  `json:",omitempty" toml:",omitempty"`
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
		TM:       DefaultTMConfig(),
		Profiler: DefaultProfilerConfig(),
		GRPC:     DefaultGRPCConfig(),
		Metrics:  DefaultMetricsConfig(),
	}
}

func DefaultTMConfig() *ServerConfig {
	return &ServerConfig{
		Enabled:       true,
		ListenAddress: "tcp://localhost:26658",
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

func DefaultMetricsConfig() *MetricsConfig {
	return &MetricsConfig{
		Enabled:         false,
		ListenAddress:   "tcp://localhost:9102",
		MetricsPath:     "/metrics",
		BlockSampleSize: 100,
	}
}
