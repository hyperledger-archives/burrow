package pipe

import (
	cfg "github.com/tendermint/go-config"
	"github.com/tendermint/log15"
)

var log = log15.New("module", "eris/erisdb_pipe")
var config cfg.Config

func init() {
	cfg.OnConfig(func(newConfig cfg.Config) {
		config = newConfig
	})
}
