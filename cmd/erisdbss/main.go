package main

import (
	ess "github.com/eris-ltd/eris-db/erisdb/erisdbss"
	"github.com/eris-ltd/eris-db/server"
	"github.com/gin-gonic/gin"
	"os"
	"path"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	baseDir := path.Join(os.TempDir(), "/.edbservers")
	ss := ess.NewServerServer(baseDir)
	proc := server.NewServeProcess(nil, ss)
	err := proc.Start()
	if err != nil {
		panic(err.Error())
	}
	<-proc.StopEventChannel()
	os.RemoveAll(baseDir)
}
