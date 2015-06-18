package web_api

import (
	"github.com/gin-gonic/gin"
	"github.com/tendermint/log15"
	"os"
	"runtime"
)

const SERVER_DURATION = 10

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log15.Root().SetHandler(log15.LvlFilterHandler(
		log15.LvlWarn,
		log15.StreamHandler(os.Stdout, log15.TerminalFormat()),
	))
	gin.SetMode(gin.ReleaseMode)
}
