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

import (
	"fmt"
	"math"

	viper "github.com/spf13/viper"
)

type (
	ServerConfig struct {
		ChainId    string
		Bind       Bind      `toml:"bind"`
		TLS        TLS       `toml:"TLS"`
		CORS       CORS      `toml:"CORS"`
		HTTP       HTTP      `toml:"HTTP"`
		WebSocket  WebSocket `toml:"web_socket"`
		Tendermint Tendermint
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

	Tendermint struct {
		RpcLocalAddress string
		Endpoint        string
	}
)

func ReadServerConfig(viper *viper.Viper) (*ServerConfig, error) {
	// TODO: [ben] replace with a more elegant way of asserting
	// the possible conversion to the type domain

	// check domain range for bind.port
	bindPortInt := viper.GetInt("bind.port")
	var bindPortUint16 uint16 = 0
	if bindPortInt >= 0 && bindPortInt <= math.MaxUint16 {
		bindPortUint16 = uint16(bindPortInt)
	} else {
		return nil, fmt.Errorf("Failed to read binding port from configuration: %v",
			bindPortInt)
	}
	// check domain range for cors.max_age
	maxAge := viper.GetInt("cors.max_age")
	var maxAgeUint64 uint64 = 0
	if maxAge >= 0 {
		maxAgeUint64 = uint64(maxAge)
	} else {
		return nil, fmt.Errorf("Failed to read maximum age for CORS: %v", maxAge)
	}
	// check domain range for websocket.max_sessions
	maxWebsocketSessions := viper.GetInt("websocket.max_sessions")
	var maxWebsocketSessionsUint16 uint16 = 0
	if maxWebsocketSessions >= 0 && maxWebsocketSessions <= math.MaxUint16 {
		maxWebsocketSessionsUint16 = uint16(maxWebsocketSessions)
	} else {
		return nil, fmt.Errorf("Failed to read maximum websocket sessions: %v",
			maxWebsocketSessions)
	}
	// check domain range for websocket.read_buffer_size
	readBufferSize := viper.GetInt("websocket.read_buffer_size")
	var readBufferSizeUint64 uint64 = 0
	if readBufferSize >= 0 {
		readBufferSizeUint64 = uint64(readBufferSize)
	} else {
		return nil, fmt.Errorf("Failed to read websocket read buffer size: %v",
			readBufferSize)
	}

	// check domain range for websocket.write_buffer_size
	writeBufferSize := viper.GetInt("websocket.read_buffer_size")
	var writeBufferSizeUint64 uint64 = 0
	if writeBufferSize >= 0 {
		writeBufferSizeUint64 = uint64(writeBufferSize)
	} else {
		return nil, fmt.Errorf("Failed to read websocket write buffer size: %v",
			writeBufferSize)
	}

	return &ServerConfig{
		Bind: Bind{
			Address: viper.GetString("bind.address"),
			Port:    bindPortUint16,
		},
		TLS: TLS{
			TLS:      viper.GetBool("tls.tls"),
			CertPath: viper.GetString("tls.cert_path"),
			KeyPath:  viper.GetString("tls.key_path"),
		},
		CORS: CORS{
			Enable:           viper.GetBool("cors.enable"),
			AllowOrigins:     viper.GetStringSlice("cors.allow_origins"),
			AllowCredentials: viper.GetBool("cors.allow_credentials"),
			AllowMethods:     viper.GetStringSlice("cors.allow_methods"),
			AllowHeaders:     viper.GetStringSlice("cors.allow_headers"),
			ExposeHeaders:    viper.GetStringSlice("cors.expose_headers"),
			MaxAge:           maxAgeUint64,
		},
		HTTP: HTTP{
			JsonRpcEndpoint: viper.GetString("http.json_rpc_endpoint"),
		},
		WebSocket: WebSocket{
			WebSocketEndpoint:    viper.GetString("websocket.endpoint"),
			MaxWebSocketSessions: maxWebsocketSessionsUint16,
			ReadBufferSize:       readBufferSizeUint64,
			WriteBufferSize:      writeBufferSizeUint64,
		},
		Tendermint: Tendermint{
			RpcLocalAddress: viper.GetString("tendermint.rpc_local_address"),
			Endpoint:        viper.GetString("tendermint.endpoint"),
		},
	}, nil
}

// NOTE: [ben] only preserved for /test/server tests; but should not be used and
// will be deprecated.
func DefaultServerConfig() *ServerConfig {
	cp := ""
	kp := ""
	return &ServerConfig{
		Bind: Bind{
			Address: "",
			Port:    1337,
		},
		TLS: TLS{TLS: false,
			CertPath: cp,
			KeyPath:  kp,
		},
		CORS: CORS{},
		HTTP: HTTP{JsonRpcEndpoint: "/rpc"},
		WebSocket: WebSocket{
			WebSocketEndpoint:    "/socketrpc",
			MaxWebSocketSessions: 50,
			ReadBufferSize:       4096,
			WriteBufferSize:      4096,
		},
		Tendermint: Tendermint{
			RpcLocalAddress: "0.0.0.0:46657",
			Endpoint:        "/websocket",
		},
	}
}
