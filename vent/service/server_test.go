// +build integration

package service_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/hyperledger/burrow/vent/config"
	"github.com/hyperledger/burrow/vent/logger"
	"github.com/hyperledger/burrow/vent/service"
	"github.com/hyperledger/burrow/vent/sqlsol"
	"github.com/hyperledger/burrow/vent/test"
	"github.com/hyperledger/burrow/vent/types"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	// run consumer to listen to events
	cfg := config.DefaultFlags()

	// create test db
	db, closeDB := test.NewTestDB(t, cfg)
	defer closeDB()

	cfg.DBSchema = db.Schema
	cfg.SpecFileOrDir = os.Getenv("GOPATH") + "/src/github.com/hyperledger/burrow/vent/test/sqlsol_example.json"
	cfg.AbiFileOrDir = os.Getenv("GOPATH") + "/src/github.com/hyperledger/burrow/vent/test/EventsTest.abi"
	cfg.GRPCAddr = testConfig.RPC.GRPC.ListenAddress

	log := logger.NewLogger(cfg.LogLevel)
	consumer := service.NewConsumer(cfg, log, make(chan types.EventData))

	projection, err := sqlsol.SpecLoader(cfg.SpecFileOrDir, false)
	abiSpec, err := sqlsol.AbiLoader(cfg.AbiFileOrDir)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		err := consumer.Run(projection, abiSpec, true)
		require.NoError(t, err)

		wg.Done()
	}()

	time.Sleep(2 * time.Second)

	// setup test server
	server := service.NewServer(cfg, log, consumer)

	httpServer := httptest.NewServer(server)
	defer httpServer.Close()

	// call health endpoint should return OK
	healthURL := fmt.Sprintf("%s/health", httpServer.URL)

	resp, err := http.Get(healthURL)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// shutdown consumer and wait for its end
	consumer.Shutdown()
	wg.Wait()

	// call health endpoint again should return error
	resp, err = http.Get(healthURL)
	require.NoError(t, err)
	require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
}
