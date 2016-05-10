package config

import (
	"github.com/eris-ltd/eris-db/files"
	"github.com/naoina/toml"
)

// Standard configuration file for the server.
type (
	ErisDBConfig struct {
		DB         DB           `toml:"db"`
		TMSP       TMSP         `toml:"tmsp"`
		Tendermint Tendermint   `toml:"tendermint"`
		Server     ServerConfig `toml":server"`
	}

	DB struct {
		Backend string `toml:"backend"`
	}

	TMSP struct {
		Listener string `toml:"listener"`
	}

	Tendermint struct {
		Host string `toml:"host"`
	}

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
		MaxWebSocketSessions uint   `toml:"max_websocket_sessions"`
		ReadBufferSize       uint   `toml:"read_buffer_size"`
		WriteBufferSize      uint   `toml:"write_buffer_size"`
	}

	Logging struct {
		ConsoleLogLevel string `toml:"console_log_level"`
		FileLogLevel    string `toml:"file_log_level"`
		LogFile         string `toml:"log_file"`
		VMLog           bool   `toml:"vm_log"`
	}
)

func DefaultServerConfig() ServerConfig {
	cp := ""
	kp := ""
	return ServerConfig{
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

func DefaultErisDBConfig() *ErisDBConfig {
	return &ErisDBConfig{
		DB: DB{
			Backend: "leveldb",
		},
		TMSP: TMSP{
			Listener: "tcp://0.0.0.0:46658",
		},
		Tendermint: Tendermint{
			Host: "0.0.0.0:46657",
		},
		Server: DefaultServerConfig(),
	}
}

// Read a TOML server configuration file.
func ReadErisDBConfig(filePath string) (*ErisDBConfig, error) {
	bts, err := files.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	cfg := &ErisDBConfig{}
	err2 := toml.Unmarshal(bts, cfg)
	if err2 != nil {
		return nil, err2
	}
	return cfg, nil
}

// Write a server configuration file.
func WriteErisDBConfig(filePath string, cfg *ErisDBConfig) error {
	bts, err := toml.Marshal(*cfg)
	if err != nil {
		return err
	}
	return files.WriteAndBackup(filePath, bts)
}

// Write a server configuration file.
func WriteServerConfig(filePath string, cfg *ServerConfig) error {
	bts, err := toml.Marshal(*cfg)
	if err != nil {
		return err
	}
	return files.WriteAndBackup(filePath, bts)
}
