package server

import (
	"github.com/eris-ltd/erisdb/files"
	"github.com/BurntSushi/toml"
	"bytes"
	"os"
	"path"
)

// TODO fix his before monday (june 8th).

// Standard configuration file for the server.
type ServerConfig struct {
	Address              string   `json:"address"`
	Port                 uint16   `json:"port"`
	CheckOrigin          bool     `json:"check_origin"`
	AllowedOrigins       []string `json:"allow_origins"`
	TLS                  bool     `json:"tls"`
	CertPath             string   `json:"cert_path"`
	KeyPath              string   `json:"key_path"`
	WebSocketPath        string   `json:"websocket_path"`
	JsonRpcPath          string   `json:"json_rpc_path"`
	MaxWebSocketSessions uint     `json:"max_websocket_sessions"`
}

func DefaultServerConfig() *ServerConfig {
	pwd, _ := os.Getwd()
	cp := path.Join(pwd, "cert.pem")
	kp := path.Join(pwd, "key.pem")
	return &ServerConfig{
		Address:              "0.0.0.0",
		Port:                 1337,
		CheckOrigin:          false,
		AllowedOrigins:       []string{"*"},
		TLS:                  false,
		CertPath:             cp,
		KeyPath:              kp,
		WebSocketPath:        "/socketrpc",
		JsonRpcPath:          "/rpc",
		MaxWebSocketSessions: 50,
	}
}

// Read a TOML server configuration file.
func ReadServerConfig(filePath string) (*ServerConfig, error) {
	cfg := &ServerConfig{}
	_, errD := toml.DecodeFile(filePath, cfg)
	if errD != nil {
		return nil, errD
	}
	return cfg, nil
}

// We want to do this with the files package methods, so writing to a byte buffer. 
func WriteServerConfig(filePath string, cfg *ServerConfig) error {
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	err := enc.Encode(cfg)
	if err != nil {
		return err
	}
	return files.WriteFileRW(filePath, buf.Bytes())
}
