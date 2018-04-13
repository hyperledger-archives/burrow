// Copyright 2017 Monax Industries Limited
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

package server

type (
	ServerConfig struct {
		ChainId   string
		Bind      Bind      `toml:"bind"`
		TLS       TLS       `toml:"TLS"`
		CORS      CORS      `toml:"CORS"`
		HTTP      HTTP      `toml:"HTTP"`
		WebSocket WebSocket `toml:"web_socket"`
	}

	Bind struct {
		Address string `toml:"address"`
		Port    uint16 `toml:"port"`
	}

	TLS struct {
		TLS      bool   `toml:"tls"`
		CertPath string `toml:"cert_path"`
		KeyPath  string `toml:"key_path"`
	}

	// Options stores configurations
	CORS struct {
		Enable           bool     `toml:"enable"`
		AllowOrigins     []string `toml:"allow_origins"`
		AllowCredentials bool     `toml:"allow_credentials"`
		AllowMethods     []string `toml:"allow_methods"`
		AllowHeaders     []string `toml:"allow_headers"`
		ExposeHeaders    []string `toml:"expose_headers"`
		MaxAge           uint64   `toml:"max_age"`
	}

	HTTP struct {
		JsonRpcEndpoint string `toml:"json_rpc_endpoint"`
	}

	WebSocket struct {
		WebSocketEndpoint    string `toml:"websocket_endpoint"`
		MaxWebSocketSessions uint16 `toml:"max_websocket_sessions"`
		ReadBufferSize       uint64 `toml:"read_buffer_size"`
		WriteBufferSize      uint64 `toml:"write_buffer_size"`
	}
)

func DefaultServerConfig() *ServerConfig {
	cp := ""
	kp := ""
	return &ServerConfig{
		Bind: Bind{
			Address: "localhost",
			Port:    1337,
		},
		TLS: TLS{TLS: false,
			CertPath: cp,
			KeyPath:  kp,
		},
		CORS: CORS{},
		HTTP: HTTP{
			JsonRpcEndpoint: "/rpc",
		},
		WebSocket: WebSocket{
			WebSocketEndpoint:    "/socketrpc",
			MaxWebSocketSessions: 50,
			ReadBufferSize:       4096,
			WriteBufferSize:      4096,
		},
	}
}
