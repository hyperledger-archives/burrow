// +build integration

package sqldb_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/lib/pq"

	"github.com/hyperledger/burrow/vent/sqldb/adapters"
	"github.com/hyperledger/burrow/vent/types"

	"github.com/hyperledger/burrow/vent/test"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger/burrow/vent/config"
)

func TestPostgresSynchronizeDB(t *testing.T) {
	testSynchronizeDB(t, config.DefaultPostgresFlags())
}

func TestPostgresCleanDB(t *testing.T) {
	testCleanDB(t, config.DefaultPostgresFlags())
}

func TestPostgresSetBlock(t *testing.T) {
	testSetBlock(t, config.DefaultPostgresFlags())
}

func TestPostgresBlockNotification(t *testing.T) {
	cfg := config.DefaultPostgresFlags()
	db, closeDB := test.NewTestDB(t, cfg)
	defer closeDB()

	errp := db.Ping()
	require.NoError(t, errp)

	functionName := "notify_height"
	channelName := "height_notification"
	pad := db.DBAdapter.(*adapters.PostgresAdapter)

	query := pad.CreateNotifyFunctionQuery(functionName, channelName, types.SQLColumnLabelHeight)
	_, err := db.DB.Exec(query)
	require.NoError(t, err)

	query = pad.CreateTriggerQuery("notify_height_trigger", types.SQLLogTableName, functionName)
	_, err = db.DB.Exec(query)
	require.NoError(t, err)

	listener := pq.NewListener(cfg.DBURL, time.Second, time.Second*20, func(event pq.ListenerEventType, err error) {
		fmt.Println(event, err)
	})
	err = listener.Listen(channelName)
	require.NoError(t, err)

	//func(event pq.ListenerEventType, err error) {
	//fmt.Printf("got event %v, error:  %v\n", event, err)
	//})

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		for n := range listener.NotificationChannel() {
			fmt.Println(n.Extra)
		wg.Done()
		return
		}
	}()

	// new
	str, dat := getBlock()
	err = db.SetBlock(str, dat)
	require.NoError(t, err)

	// read
	_, err = db.GetLastBlockID()
	require.NoError(t, err)

	_, err = db.GetBlock(dat.Block)
	require.NoError(t, err)

	wg.Wait()
}
