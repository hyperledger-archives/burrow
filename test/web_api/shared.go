package web_api

import (
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/gin-gonic/gin"
	"github.com/eris-ltd/eris-db/tendermint/log15"
	"os"
	"runtime"
)

const SERVER_DURATION = 10

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log15.Root().SetHandler(log15.LvlFilterHandler(
		log15.LvlInfo,
		log15.StreamHandler(os.Stdout, log15.TerminalFormat()),
	))
	gin.SetMode(gin.ReleaseMode)
}
