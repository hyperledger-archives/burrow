// +build integration

package sqldb_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/hyperledger/burrow/vent/sqldb/adapters"
	"github.com/hyperledger/burrow/vent/test"
	"github.com/hyperledger/burrow/vent/types"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

var tables = types.DefaultSQLTableNames
var columns = types.DefaultSQLColumnNames

func TestPostgresSynchronizeDB(t *testing.T) {
	testSynchronizeDB(t, test.PostgresVentConfig(""))
}

func TestPostgresCleanDB(t *testing.T) {
	testCleanDB(t, test.PostgresVentConfig(""))
}

func TestPostgresSetBlock(t *testing.T) {
	testSetBlock(t, test.PostgresVentConfig(""))
}

func TestRestore(t *testing.T) {
	testRestore(t, test.PostgresVentConfig(""))
}

func TestPostgresBlockNotification(t *testing.T) {
	cfg := test.PostgresVentConfig("")
	db, closeDB := test.NewTestDB(t, cfg)
	defer closeDB()

	errp := db.Ping()
	require.NoError(t, errp)

	functionName := "notify_height"
	channelName := "height_notification"
	pad := db.DBAdapter.(*adapters.PostgresAdapter)

	for i := 0; i < 2; i++ {
		query := pad.CreateNotifyFunctionQuery(functionName, channelName, columns.Height)
		_, err := db.DB.Exec(query)
		require.NoError(t, err)

		query = pad.CreateTriggerQuery("notify_height_trigger", tables.Log, functionName)
		_, err = db.DB.Exec(query)
		require.NoError(t, err)
	}

	listener := pq.NewListener(cfg.DBURL, time.Second, time.Second*20, func(event pq.ListenerEventType, err error) {
		require.NoError(t, err)
	})
	err := listener.Listen(channelName)
	require.NoError(t, err)

	// new block
	str, dat := getBlock()

	errCh := make(chan error)
	go func() {
		type payload struct {
			Height string `json:"_height"`
		}
		for n := range listener.NotificationChannel() {
			pl := new(payload)
			err := json.Unmarshal([]byte(n.Extra), pl)
			if err != nil {
				errCh <- err
				return
			}
			if pl.Height != "" {
				if strconv.FormatUint(dat.BlockHeight, 10) != pl.Height {
					errCh <- fmt.Errorf("got height %s from notification but expected %d",
						pl.Height, dat.BlockHeight)
				}
				errCh <- nil
				return
			}
		}
	}()

	// Set it
	err = db.SetBlock(test.ChainID, str, dat)
	require.NoError(t, err)

	// read
	_, err = db.LastBlockHeight(test.ChainID)
	require.NoError(t, err)

	_, err = db.GetBlock(test.ChainID, dat.BlockHeight)
	require.NoError(t, err)

	require.NoError(t, <-errCh)
}
