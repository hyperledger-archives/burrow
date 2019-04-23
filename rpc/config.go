package rpc

// 'LocalHost' gets interpreted as ipv6
// TODO: revisit this
const LocalHost = "127.0.0.1"

type RPCConfig struct {
	Info     *ServerConfig  `json:",omitempty" toml:",omitempty"`
	Profiler *ServerConfig  `json:",omitempty" toml:",omitempty"`
	GRPC     *ServerConfig  `json:",omitempty" toml:",omitempty"`
	Metrics  *MetricsConfig `json:",omitempty" toml:",omitempty"`
}

type ServerConfig struct {
	Enabled    bool
	ListenHost string
	ListenPort string
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
	}
}

func DefaultInfoConfig() *ServerConfig {
	return &ServerConfig{
		Enabled:    true,
		ListenHost: LocalHost,
		ListenPort: "26658",
	}
}

func DefaultGRPCConfig() *ServerConfig {
	return &ServerConfig{
		Enabled:    true,
		ListenHost: LocalHost,
		ListenPort: "10997",
	}
}

func DefaultProfilerConfig() *ServerConfig {
	return &ServerConfig{
		Enabled:    false,
		ListenHost: LocalHost,
		ListenPort: "6060",
	}
}

func DefaultMetricsConfig() *MetricsConfig {
	return &MetricsConfig{
		ServerConfig: ServerConfig{
			Enabled:    false,
			ListenHost: LocalHost,
			ListenPort: "9102",
		},
		MetricsPath:     "/metrics",
		BlockSampleSize: 100,
	}
}
