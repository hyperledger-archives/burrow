// Copyright 2015, 2016 Eris Industries (UK) Ltd.
// This file is part of Eris-RT

// Eris-RT is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Eris-RT is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Eris-RT.  If not, see <http://www.gnu.org/licenses/>.

package server

import (
  "fmt"
  "math"

  viper "github.com/spf13/viper"
)

type (
  ServerConfig struct {
    Bind      Bind      `toml:"bind"`
    TLS       TLS       `toml:"TLS"`
    CORS      CORS      `toml:"CORS"`
    HTTP      HTTP      `toml:"HTTP"`
    WebSocket WebSocket `toml:"web_socket"`
    Logging   Logging   `toml:"logging"`
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

  Logging struct {
    ConsoleLogLevel string `toml:"console_log_level"`
    FileLogLevel    string `toml:"file_log_level"`
    LogFile         string `toml:"log_file"`
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

  return &ServerConfig {
    Bind: Bind {
      Address: viper.GetString("bind.address"),
      Port:    bindPortUint16,
    },
    TLS: TLS {
      TLS:      viper.GetBool("tls.tls"),
      CertPath: viper.GetString("tls.cert_path"),
      KeyPath:  viper.GetString("tls.key_path"),
    },
    CORS: CORS {
      Enable:           viper.GetBool("cors.enable"),
      AllowOrigins:     viper.GetStringSlice("cors.allow_origins"),
      AllowCredentials: viper.GetBool("cors.allow_credentials"),
      AllowMethods:     viper.GetStringSlice("cors.allow_methods"),
      AllowHeaders:     viper.GetStringSlice("cors.allow_headers"),
      ExposeHeaders:    viper.GetStringSlice("cors.expose_headers"),
      MaxAge:           maxAgeUint64,
    },
    HTTP: HTTP {
      JsonRpcEndpoint: viper.GetString("http.json_rpc_endpoint"),
    },
    WebSocket: WebSocket {
      WebSocketEndpoint:    viper.GetString("websocket.endpoint"),
      MaxWebSocketSessions: maxWebsocketSessionsUint16,
      ReadBufferSize:       readBufferSizeUint64,
      WriteBufferSize:      writeBufferSizeUint64,
    },
    Logging: Logging{
      ConsoleLogLevel: viper.GetString("logging.console_log_level"),
      FileLogLevel:    viper.GetString("logging.file_log_level"),
      LogFile:         viper.GetString("logging.log_file"),
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
		Logging: Logging{
			ConsoleLogLevel: "info",
			FileLogLevel:    "warn",
			LogFile:         "",
		},
	}
}
