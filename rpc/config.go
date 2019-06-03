// Copyright 2019 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
