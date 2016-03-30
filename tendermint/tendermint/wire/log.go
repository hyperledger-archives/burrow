package wire

import (
	"github.com/eris-ltd/eris-db/tendermint/log15"
	"github.com/eris-ltd/eris-db/tendermint/tendermint/logger"
)

var log = logger.New("module", "binary")

func init() {
	log.SetHandler(
		log15.LvlFilterHandler(
			log15.LvlWarn,
			//log15.LvlDebug,
			logger.RootHandler(),
		),
	)
}
