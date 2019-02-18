// +build integration

package service_test

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hyperledger/burrow/vent/types"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger/burrow/vent/test"
)

func TestPostgresConsumer(t *testing.T) {
	testConsumer(t, test.PostgresVentConfig())
}

func TestPostgresInvalidUTF8(t *testing.T) {
	testInvalidUTF8(t, test.PostgresVentConfig())
}

func TestPostgresDeleteEvent(t *testing.T) {
	testDeleteEvent(t, test.PostgresVentConfig())
}

func TestPostgresResume(t *testing.T) {
	testResume(t, test.PostgresVentConfig())
}

func TestPostgresTriggers(t *testing.T) {
	tCli := test.NewTransactClient(t, testConfig.RPC.GRPC.ListenAddress)
	create := test.CreateContract(t, tCli, inputAccount.GetAddress())

	// generate events
	name := "TestTriggerEvent"
	description := "Trigger it!"
	txe := test.CallAddEvent(t, tCli, inputAccount.GetAddress(), create.Receipt.ContractAddress, name, description)

	cfg := test.PostgresVentConfig()
	// create test db
	_, closeDB := test.NewTestDB(t, cfg)
	defer closeDB()

	// Create a postgres notification listener
	listener := pq.NewListener(cfg.DBURL, time.Second, time.Second*20, func(event pq.ListenerEventType, err error) {
		require.NoError(t, err)
	})

	// These are defined n sqlsol_example.json
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
	runConsumer(t, cfg)

	// Give events a chance
	time.Sleep(time.Second)
	// Assert we get expected returns
	t.Logf("latest height: %d, txe height: %d", height, txe.Height)
	assert.True(t, height >= txe.Height)
	assert.Equal(t, `{"_action" : "INSERT", "testdescription" : "\\x5472696767657220697421000000000100000000000000000000000000000000", "testname" : "TestTriggerEvent"}`, notifications["meta"])
	assert.Equal(t, `{"_action" : "INSERT", "testdescription" : "\\x5472696767657220697421000000000100000000000000000000000000000000", "testkey" : "\\x544553545f4556454e5453000000000000000000000000000000000000000000", "testname" : "TestTriggerEvent"}`,
		notifications["keyed_meta"])
}
