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

	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/integration"
	"github.com/hyperledger/burrow/integration/rpctest"
	"github.com/hyperledger/burrow/vent/config"
	"github.com/hyperledger/burrow/vent/logger"
	"github.com/hyperledger/burrow/vent/service"
	"github.com/hyperledger/burrow/vent/sqlsol"
	"github.com/hyperledger/burrow/vent/test"
	"github.com/hyperledger/burrow/vent/types"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	kern, shutdown := integration.RunNode(t, rpctest.GenesisDoc, rpctest.PrivateAccounts)
	defer shutdown()
	t.Parallel()

	t.Run("Group", func(t *testing.T) {
		t.Run("Run", func(t *testing.T) {
			// run consumer to listen to events
			cfg := config.DefaultVentConfig()

			// create test db
			_, closeDB := test.NewTestDB(t, kern.Blockchain.ChainID(), cfg)
			defer closeDB()

			cfg.SpecFileOrDirs = []string{os.Getenv("GOPATH") + "/src/github.com/hyperledger/burrow/vent/test/sqlsol_example.json"}
			cfg.AbiFileOrDirs = []string{os.Getenv("GOPATH") + "/src/github.com/hyperledger/burrow/vent/test/EventsTest.abi"}
			cfg.GRPCAddr = kern.GRPCListenAddress().String()

			log := logger.NewLogger(cfg.LogLevel)
			consumer := service.NewConsumer(cfg, log, make(chan types.EventData))

			projection, err := sqlsol.SpecLoader(cfg.SpecFileOrDirs, false)
			abiSpec, err := abi.LoadPath(cfg.AbiFileOrDirs...)

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
		})
	})
}
