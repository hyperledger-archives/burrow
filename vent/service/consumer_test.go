// +build integration

package service_test

import (
	"math/rand"
	"path"
	"runtime"
	"testing"
	"time"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/rpc/rpctransact"

	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/vent/config"
	"github.com/hyperledger/burrow/vent/logger"
	"github.com/hyperledger/burrow/vent/service"
	"github.com/hyperledger/burrow/vent/sqlsol"
	"github.com/hyperledger/burrow/vent/test"
	"github.com/hyperledger/burrow/vent/types"

	"github.com/stretchr/testify/require"
)

func testConsumer(t *testing.T, chainid string, cfg *config.VentConfig, tcli rpctransact.TransactClient, inputAddress crypto.Address) {
	create := test.CreateContract(t, tcli, inputAddress)

	// generate events
	name := "TestEvent1"
	description := "Description of TestEvent1"
	txeA := test.CallAddEvent(t, tcli, inputAddress, create.Receipt.ContractAddress, name, description)

	name = "TestEvent2"
	description = "Description of TestEvent2"
	test.CallAddEvent(t, tcli, inputAddress, create.Receipt.ContractAddress, name, description)

	name = "TestEvent3"
	description = "Description of TestEvent3"
	test.CallAddEvent(t, tcli, inputAddress, create.Receipt.ContractAddress, name, description)

	name = "TestEvent4"
	description = "Description of TestEvent4"
	txeB := test.CallAddEvent(t, tcli, inputAddress, create.Receipt.ContractAddress, name, description)

	// create test db
	db, closeDB := test.NewTestDB(t, chainid, cfg)
	defer closeDB()

	// Run the consumer
	runConsumer(t, cfg)

	// test data stored in database for two different block ids
	eventColumnName := "EventTest"

	blockID := txeA.Height
	eventData, err := db.GetBlock(blockID)

	require.NoError(t, err)
	require.Equal(t, blockID, eventData.BlockHeight)
	require.Equal(t, 3, len(eventData.Tables))

	tblData := eventData.Tables[eventColumnName]
	require.Equal(t, 1, len(tblData))
	require.Equal(t, "LogEvent", tblData[0].RowData["_eventtype"].(string))
	require.Equal(t, "UpdateTestEvents", tblData[0].RowData["_eventname"].(string))

	blockID = txeB.Height
	eventData, err = db.GetBlock(blockID)
	require.NoError(t, err)
	require.Equal(t, blockID, eventData.BlockHeight)
	require.Equal(t, 3, len(eventData.Tables))

	tblData = eventData.Tables[eventColumnName]
	require.Equal(t, 1, len(tblData))
	require.Equal(t, "LogEvent", tblData[0].RowData["_eventtype"].(string))
	require.Equal(t, "UpdateTestEvents", tblData[0].RowData["_eventname"].(string))

	// block & tx raw data also persisted
	if cfg.DBBlockTx {
		tblData = eventData.Tables[types.SQLBlockTableName]
		require.Equal(t, 1, len(tblData))

		tblData = eventData.Tables[types.SQLTxTableName]
		require.Equal(t, 1, len(tblData))
		require.Equal(t, txeB.TxHash.String(), tblData[0].RowData["_txhash"].(string))
	}

	//Restore
	ti := time.Now().Local().AddDate(10, 0, 0)
	err = db.RestoreDB(ti, "RESTORED")
	require.NoError(t, err)
}

func testDeleteEvent(t *testing.T, chainid string, cfg *config.VentConfig, tcli rpctransact.TransactClient, inputAddress crypto.Address) {
	create := test.CreateContract(t, tcli, inputAddress)

	// create test db
	db, closeDB := test.NewTestDB(t, chainid, cfg)
	defer closeDB()

	// test data stored in database for two different block ids
	eventColumnName := "EventTest"

	// Add a test event
	name := "TestEventForDeletion"
	description := "to be deleted"
	txeAdd := test.CallAddEvent(t, tcli, inputAddress, create.Receipt.ContractAddress, name, description)

	// Spin the consumer
	runConsumer(t, cfg)

	// Expect block table, tx table, and EventTest table
	eventData, err := db.GetBlock(txeAdd.Height)
	require.NoError(t, err)
	require.Equal(t, txeAdd.Height, eventData.BlockHeight)
	require.Equal(t, 3, len(eventData.Tables))

	// Expect data in the EventTest table
	tblData := eventData.Tables[eventColumnName]
	require.Equal(t, 1, len(tblData))
	require.Equal(t, "LogEvent", tblData[0].RowData["_eventtype"].(string))
	require.Equal(t, "UpdateTestEvents", tblData[0].RowData["_eventname"].(string))

	// Now emit a deletion event for that table
	test.CallRemoveEvent(t, tcli, inputAddress, create.Receipt.ContractAddress, name)
	runConsumer(t, cfg)

	eventData, err = db.GetBlock(txeAdd.Height)
	require.NoError(t, err)
	require.Equal(t, txeAdd.Height, eventData.BlockHeight)
	require.Equal(t, 3, len(eventData.Tables))

	// Check the row was deleted
	tblData = eventData.Tables[eventColumnName]
	require.Equal(t, 0, len(tblData))
}

func testResume(t *testing.T, chainid string, cfg *config.VentConfig) {
	_, closeDB := test.NewTestDB(t, chainid, cfg)
	defer closeDB()

	numRestarts := 6
	// Add some pseudo-random timings
	rnd := rand.New(rand.NewSource(4634653))
	time.Sleep(time.Second)
	var expectedHeight uint64
	for i := 0; i < numRestarts; i++ {
		// wait up to a second
		time.Sleep(time.Millisecond * time.Duration(rnd.Int63n(1000)))
		for ed := range runConsumer(t, cfg) {
			expectedHeight++
			t.Logf("expecting block: %d, got block: %d", expectedHeight, ed.BlockHeight)
			require.Equal(t, expectedHeight, ed.BlockHeight, "should get monotonic sequential sequence")
		}
	}
}

func testInvalidUTF8(t *testing.T, chainid string, cfg *config.VentConfig, tcli rpctransact.TransactClient, inputAddress crypto.Address) {
	create := test.CreateContract(t, tcli, inputAddress)

	// The code point for ó is less than 255 but needs two unicode bytes - it's value expressed as a single byte
	// is in the private use area so is invalid.
	goodString := "Cliente - Doc. identificación"

	// generate events
	name := service.BadStringToHexFunction(goodString)
	description := "Description of TestEvent1"
	test.CallAddEvent(t, tcli, inputAddress, create.Receipt.ContractAddress, name, description)

	// create test db
	_, closeDB := test.NewTestDB(t, chainid, cfg)
	defer closeDB()

	// Run the consumer with this event - this used to create an error on UPSERT
	runConsumer(t, cfg)

	// Error we used to get before fixing this test case:
	//require.Error(t, err)
	//require.Contains(t, err.Error(), "pq: invalid byte sequence for encoding \"UTF8\": 0xf3 0x6e")
}

func newConsumer(t *testing.T, cfg *config.VentConfig) *service.Consumer {
	// Resolve relative path to test dir
	_, testFile, _, _ := runtime.Caller(0)
	testDir := path.Join(path.Dir(testFile), "..", "test")

	cfg.SpecFileOrDirs = []string{path.Join(testDir, "sqlsol_example.json")}
	cfg.AbiFileOrDirs = []string{path.Join(testDir, "EventsTest.abi")}
	cfg.DBBlockTx = true

	log := logger.NewLogger(cfg.LogLevel)
	ch := make(chan types.EventData, 100)
	return service.NewConsumer(cfg, log, ch)
}

// Run consumer to listen to events
func runConsumer(t *testing.T, cfg *config.VentConfig) chan types.EventData {
	consumer := newConsumer(t, cfg)

	projection, err := sqlsol.SpecLoader(cfg.SpecFileOrDirs, cfg.DBBlockTx)
	require.NoError(t, err)

	abiSpec, err := abi.LoadPath(cfg.AbiFileOrDirs...)
	require.NoError(t, err)

	err = consumer.Run(projection, abiSpec, false)
	require.NoError(t, err)
	return consumer.EventsChannel
}
