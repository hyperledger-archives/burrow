package erisdb

import (
	cfg "github.com/tendermint/go-config"
)

var config cfg.Config

func init() {
	cfg.OnConfig(func(newConfig cfg.Config) {
		config = newConfig
	})
}
