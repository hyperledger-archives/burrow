// +build integration

package service_test

import (
	"fmt"
	"path"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hyperledger/burrow/vent/config"
	"github.com/hyperledger/burrow/vent/logger"
	"github.com/hyperledger/burrow/vent/service"
	"github.com/hyperledger/burrow/vent/sqldb"
	"github.com/hyperledger/burrow/vent/sqlsol"
	"github.com/hyperledger/burrow/vent/test"
	"github.com/hyperledger/burrow/vent/types"

	"github.com/stretchr/testify/require"
)

func TestConsumer(t *testing.T) {
	tCli := test.NewTransactClient(t, testConfig.RPC.GRPC.ListenAddress)
	create := test.CreateContract(t, tCli, inputAccount.GetAddress())

	// generate events
	name := "TestEvent1"
	description := "Description of TestEvent1"
	txe := test.CallAddEvent(t, tCli, inputAccount.GetAddress(), create.Receipt.ContractAddress, name, description)

	name = "TestEvent2"
	description = "Description of TestEvent2"
	txe = test.CallAddEvent(t, tCli, inputAccount.GetAddress(), create.Receipt.ContractAddress, name, description)

	name = "TestEvent3"
	description = "Description of TestEvent3"
	txe = test.CallAddEvent(t, tCli, inputAccount.GetAddress(), create.Receipt.ContractAddress, name, description)

	name = "TestEvent4"
	description = "Description of TestEvent4"
	txe = test.CallAddEvent(t, tCli, inputAccount.GetAddress(), create.Receipt.ContractAddress, name, description)

	// workaround for off-by-one on latest bound fixed in burrow
	time.Sleep(time.Second * 2)

	cfg := config.DefaultFlags()
	// create test db
	db, closeDB := test.NewTestDB(t, cfg)
	defer closeDB()

	err := runConsumer(db, cfg)
	require.NoError(t, err)

	// test data stored in database for two different block ids
	eventName := "EventTest"

	blockID := "2"
	eventData, err := db.GetBlock(blockID)
	require.NoError(t, err)
	require.Equal(t, "2", eventData.Block)
	require.Equal(t, 3, len(eventData.Tables))

	tblData := eventData.Tables[strings.ToLower(eventName)]
	require.Equal(t, 1, len(tblData))
	require.Equal(t, "LogEvent", tblData[0].RowData["_eventtype"].(string))
	require.Equal(t, "UpdateTestEvents", tblData[0].RowData["_eventname"].(string))

	blockID = "5"
	eventData, err = db.GetBlock(blockID)
	require.NoError(t, err)
	require.Equal(t, "5", eventData.Block)
	require.Equal(t, 3, len(eventData.Tables))

	tblData = eventData.Tables[strings.ToLower(eventName)]
	require.Equal(t, 1, len(tblData))
	require.Equal(t, "LogEvent", tblData[0].RowData["_eventtype"].(string))
	require.Equal(t, "UpdateTestEvents", tblData[0].RowData["_eventname"].(string))

	// block & tx raw data also persisted
	if cfg.DBBlockTx {
		tblData = eventData.Tables[types.SQLBlockTableName]
		require.Equal(t, 1, len(tblData))

		tblData = eventData.Tables[types.SQLTxTableName]
		require.Equal(t, 1, len(tblData))
		require.Equal(t, txe.TxHash.String(), tblData[0].RowData["_txhash"].(string))
	}

	//Restore
	ti := time.Now().Local().AddDate(10, 0, 0)
	err = db.RestoreDB(ti, "RESTORED")
	require.NoError(t, err)
}

func TestInvalidUTF8(t *testing.T) {
	tCli := test.NewTransactClient(t, testConfig.RPC.GRPC.ListenAddress)
	create := test.CreateContract(t, tCli, inputAccount.GetAddress())

	// The code point for ó is less than 255 but needs two unicode bytes - it's value expressed as a single byte
	// is in the private use area so is invalid.
	goodString := "Cliente - Doc. identificación"

	// generate events
	name := service.BadStringToHexFunction(goodString)
	fmt.Println(name)
	description := "Description of TestEvent1"
	test.CallAddEvent(t, tCli, inputAccount.GetAddress(), create.Receipt.ContractAddress, name, description)

	cfg := config.DefaultFlags()
	// create test db
	db, closeDB := test.NewTestDB(t, cfg)
	defer closeDB()

	// Run the consumer with this event - this used to create an error on UPSERT
	err := runConsumer(db, cfg)
	require.NoError(t, err)

	// Error we used to get before fixing this test case:
	//require.Error(t, err)
	//require.Contains(t, err.Error(), "pq: invalid byte sequence for encoding \"UTF8\": 0xf3 0x6e")
}

// Run consumer to listen to events
func runConsumer(db *sqldb.SQLDB, cfg *config.Flags) (err error) {
	// Resolve relative path to test dir
	_, testFile, _, _ := runtime.Caller(0)
	testDir := path.Join(path.Dir(testFile), "..", "test")

	cfg.DBSchema = db.Schema
	cfg.SpecFile = path.Join(testDir, "sqlsol_example.json")
	cfg.AbiFile = path.Join(testDir, "EventsTest.abi")
	cfg.GRPCAddr = testConfig.RPC.GRPC.ListenAddress
	cfg.DBBlockTx = true

	log := logger.NewLogger("info")
	consumer := service.NewConsumer(cfg, log, make(chan types.EventData))

	parser, err := sqlsol.SpecLoader("", cfg.SpecFile, cfg.DBBlockTx)
	if err != nil {
		return err
	}
	abiSpec, err := sqlsol.AbiLoader("", cfg.AbiFile)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = consumer.Run(parser, abiSpec, false)
	}()

	// wait for block streams to start
	time.Sleep(time.Second * 2)
	consumer.Shutdown()

	wg.Wait()
	return
}
