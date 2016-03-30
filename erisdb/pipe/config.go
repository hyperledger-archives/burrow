package pipe

import (
	"github.com/eris-ltd/eris-db/tendermint/log15"
	cfg "github.com/eris-ltd/eris-db/tendermint/tendermint/config"
)

var log = log15.New("module", "eris/erisdb_pipe")
var config cfg.Config

func init() {
	cfg.OnConfig(func(newConfig cfg.Config) {
		config = newConfig
	})
}
