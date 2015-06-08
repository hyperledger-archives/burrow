package main 

import (
	ess "github.com/eris-ltd/erisdb/erisdb/erisdbss"
	"github.com/eris-ltd/erisdb/server"
	"os"
	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	baseDir := os.Getenv("HOME") + "/.edbservers"
	ss := ess.NewServerServer(baseDir)
	proc := server.NewServeProcess(nil, ss)
	err := proc.Start()
	if err != nil {
		panic(err.Error())
	}
	<- proc.StopEventChannel()
	os.RemoveAll(baseDir)
}