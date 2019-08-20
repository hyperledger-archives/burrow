// +build integration

package service_test

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/hyperledger/burrow/integration"
	"github.com/hyperledger/burrow/integration/rpctest"

	"github.com/stretchr/testify/assert"

	"github.com/hyperledger/burrow/vent/types"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger/burrow/vent/test"
)

func TestPostgresConsumer(t *testing.T) {
	privateAccounts := rpctest.PrivateAccounts
	kern, shutdown := integration.RunNode(t, rpctest.GenesisDoc, privateAccounts)
	defer shutdown()
	inputAddress := privateAccounts[1].GetAddress()
	grpcAddress := kern.GRPCListenAddress().String()
	tcli := test.NewTransactClient(t, grpcAddress)

	t.Parallel()
	time.Sleep(2 * time.Second)

	t.Run("Group", func(t *testing.T) {
		t.Run("PostgresConsumer", func(t *testing.T) {
			testConsumer(t, kern.Blockchain.ChainID(), test.PostgresVentConfig(grpcAddress), tcli, inputAddress)
		})

		t.Run("PostgresInvalidUTF8", func(t *testing.T) {
			testInvalidUTF8(t, test.PostgresVentConfig(grpcAddress), tcli, inputAddress)
		})

		t.Run("PostgresDeleteEvent", func(t *testing.T) {
			testDeleteEvent(t, kern.Blockchain.ChainID(), test.PostgresVentConfig(grpcAddress), tcli, inputAddress)
		})

		t.Run("PostgresResume", func(t *testing.T) {
			testResume(t, test.PostgresVentConfig(grpcAddress))
		})

		t.Run("PostgresTriggers", func(t *testing.T) {
			tCli := test.NewTransactClient(t, kern.GRPCListenAddress().String())
			create := test.CreateContract(t, tCli, inputAddress)

			// generate events
			name := "TestTriggerEvent"
			description := "Trigger it!"
			txe := test.CallAddEvent(t, tCli, inputAddress, create.Receipt.ContractAddress, name, description)

			cfg := test.PostgresVentConfig(grpcAddress)
			// create test db
			_, closeDB := test.NewTestDB(t, cfg)
			defer closeDB()

			// Create a postgres notification listener
			listener := pq.NewListener(cfg.DBURL, time.Second, time.Second*20, func(event pq.ListenerEventType, err error) {
				require.NoError(t, err)
			})

			// These are defined in sqlsol_view.json
			err := listener.Listen("meta")
			require.NoError(t, err)

			err = listener.Listen("keyed_meta")
			require.NoError(t, err)

			err = listener.Listen(types.BlockHeightLabel)
			require.NoError(t, err)

			type payload struct {
				Height string `json:"_height"`
			}

			var height uint64
			notifications := make(map[string]string)
			go func() {
				for n := range listener.Notify {
					notifications[n.Channel] = n.Extra
					if n.Channel == types.BlockHeightLabel {
						pl := new(payload)
						err := json.Unmarshal([]byte(n.Extra), pl)
						if err != nil {
							panic(err)
						}
						if pl.Height != "" {
							height, err = strconv.ParseUint(pl.Height, 10, 64)
							if err != nil {
								panic(err)
							}
						}
					}
				}
			}()
			resolveSpec(cfg, testViewSpec)
			runConsumer(t, cfg)

			// Give events a chance
			time.Sleep(time.Second)
			// Assert we get expected returns
			t.Logf("latest height: %d, txe height: %d", height, txe.Height)
			assert.True(t, height >= txe.Height)
			assert.Equal(t, `{"_action" : "INSERT", "testdescription" : "\\x5472696767657220697421000000000000000000000000000000000000000000", "testname" : "TestTriggerEvent"}`, notifications["meta"])
			assert.Equal(t, `{"_action" : "INSERT", "testdescription" : "\\x5472696767657220697421000000000000000000000000000000000000000000", "testkey" : "\\x544553545f4556454e5453000000000000000000000000000000000000000000", "testname" : "TestTriggerEvent"}`,
				notifications["keyed_meta"])
		})
	})
}
