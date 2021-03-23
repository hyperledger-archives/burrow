// +build integration

package service_test

import (
	"fmt"
	"math/rand"
	"path"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/hyperledger/burrow/vent/chain/ethereum"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/logging/logconfig"
	"github.com/hyperledger/burrow/rpc/rpctransact"
	"github.com/hyperledger/burrow/vent/config"
	"github.com/hyperledger/burrow/vent/service"
	"github.com/hyperledger/burrow/vent/sqldb"
	"github.com/hyperledger/burrow/vent/sqlsol"
	"github.com/hyperledger/burrow/vent/test"
	"github.com/hyperledger/burrow/vent/types"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

const (
	testViewSpec = "sqlsol_view.json"
	testLogSpec  = "sqlsol_log.json"
)

var tables = types.DefaultSQLTableNames

// Tweak logger for debug purposes here
var logger = logconfig.Sink().Terminal().FilterScope(ethereum.Scope).LoggingConfig().WithTrace().MustLogger()

func testConsumer(t *testing.T, chainID string, cfg *config.VentConfig, tcli test.TransactClient,
	inputAddress crypto.Address) {

	create := test.CreateContract(t, tcli, inputAddress)
	eventTestTableName := "EventTest"
	require.True(t, create.Receipt.CreatesContract)
	cfg.WatchAddresses = []crypto.Address{create.Receipt.ContractAddress}

	// TODO: strengthen this test to count events in and out
	t.Run("view mode", func(t *testing.T) {
		// create test db
		db, closeDB := test.NewTestDB(t, cfg)
		defer closeDB()
		resolveSpec(cfg, testViewSpec)

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

		// Run the consumer
		runConsumer(t, cfg)

		// test data stored in database for two different block heights
		ensureEvents(t, db, chainID, eventTestTableName, txeA.Height, 1)
		eventData := ensureEvents(t, db, chainID, eventTestTableName, txeB.Height, 1)

		// block & tx raw data also persisted
		if cfg.SpecOpt&sqlsol.Block > 0 {
			tblData := eventData.Tables[tables.Block]
			require.Equal(t, 1, len(tblData))

		}
		if cfg.SpecOpt&sqlsol.Tx > 0 {
			tblData := eventData.Tables[tables.Tx]
			require.Equal(t, 1, len(tblData))
			require.Equal(t, txeB.TxHash.String(), tblData[0].RowData["_txhash"].(string))
		}

		// Restore
		err := db.RestoreDB(time.Time{}, "RESTORED")
		require.NoError(t, err)
	})

	t.Run("log mode", func(t *testing.T) {
		db, closeDB := test.NewTestDB(t, cfg)
		defer closeDB()
		resolveSpec(cfg, testLogSpec)

		name := "TestEvent5"
		description := "Description of TestEvent5"
		txeC := test.CallAddEvents(t, tcli, inputAddress, create.Receipt.ContractAddress, name, description)
		runConsumer(t, cfg)
		ensureEvents(t, db, chainID, eventTestTableName, txeC.Height, 2)
	})

	t.Run("continuity", func(t *testing.T) {
		batches := 20
		batchSize := 5
		totalTx := batches * batchSize
		db, closeDB := test.NewTestDB(t, cfg)
		defer closeDB()
		resolveSpec(cfg, testLogSpec)

		receipts := make(chan *exec.TxExecution, totalTx)
		wg := new(sync.WaitGroup)
		wg.Add(totalTx)

		for i := 0; i < batches; i++ {
			for j := 0; j < batchSize; j++ {
				name := fmt.Sprintf("Continuity_%d_%d", i, j)
				go func() {
					receipts <- test.CallAddEvent(t, tcli, inputAddress, create.Receipt.ContractAddress, name, "Blah")
					wg.Done()
				}()
			}
		}

		wg.Wait()
		close(receipts)
		runConsumer(t, cfg)

		txeByHeight := make(map[uint64][]*exec.TxExecution)
		for txe := range receipts {
			txeByHeight[txe.Height] = append(txeByHeight[txe.Height], txe)
		}

		for height, txes := range txeByHeight {
			ensureEvents(t, db, chainID, eventTestTableName, height, uint64(len(txes)))
		}

	})

}

func testDeleteEvent(t *testing.T, chainID string, cfg *config.VentConfig, tcli rpctransact.TransactClient, inputAddress crypto.Address) {
	create := test.CreateContract(t, tcli, inputAddress)

	eventColumnName := "EventTest"
	name := "TestEventForDeletion"
	description := "to be deleted"

	// test data stored in database for two different block ids

	// create test db
	db, closeDB := test.NewTestDB(t, cfg)
	defer closeDB()
	resolveSpec(cfg, testViewSpec)

	// Add a test event
	txeAdd := test.CallAddEvent(t, tcli, inputAddress, create.Receipt.ContractAddress, name, description)

	// Spin the consumer
	runConsumer(t, cfg)

	// Expect block table, tx table, and EventTest table
	ensureEvents(t, db, chainID, eventColumnName, txeAdd.Height, 1)

	// Now emit a deletion event for that table
	test.CallRemoveEvent(t, tcli, inputAddress, create.Receipt.ContractAddress, name)
	runConsumer(t, cfg)
	ensureEvents(t, db, chainID, eventColumnName, txeAdd.Height, 0)

	// delete not allowed on log mode
}

func ensureEvents(t *testing.T, db *sqldb.SQLDB, chainID, table string, height, numEvents uint64) types.EventData {
	eventData, err := db.GetBlock(chainID, height)
	require.NoError(t, err)
	require.Equal(t, height, eventData.BlockHeight)
	require.Equal(t, 3, len(eventData.Tables))

	// Check the number of rows
	tblData := eventData.Tables[table]
	if !assert.Equal(t, numEvents, uint64(len(tblData))) {
		t.Fatal(logconfig.JSONString(tblData))
	}

	if numEvents > 0 && len(tblData) > 0 {
		// Expect data in the EventTest table
		require.Equal(t, "LogEvent", tblData[0].RowData["_eventtype"].(string))
		require.Equal(t, "UpdateTestEvents", tblData[0].RowData["_eventname"].(string))
		lastTxIndex := ""
		var eventIndex uint64
		for _, datum := range tblData {
			txIndex := datum.RowData["_txindex"].(string)
			if lastTxIndex != txIndex {
				eventIndex = 0
				lastTxIndex = txIndex
			}

			good := rowEqual(t, datum.RowData, "_height", height) &&
				rowEqual(t, datum.RowData, "_eventindex", eventIndex)
			if !good {
				t.Fatal(logconfig.JSONString(tblData))
			}

			assert.Equal(t, fmt.Sprintf("%d", height), datum.RowData["_height"])
			eventIndex++
		}
	} else if numEvents > 0 && len(tblData) == 0 {
		require.Failf(t, "no events found", "expected %d", numEvents)
	}

	return eventData
}

func rowEqual(t *testing.T, row map[string]interface{}, key string, expectedIndex uint64) bool {
	return assert.Equal(t, strconv.FormatUint(expectedIndex, 10), row[key].(string))
}

func testResume(t *testing.T, cfg *config.VentConfig) {
	_, closeDB := test.NewTestDB(t, cfg)
	defer closeDB()
	resolveSpec(cfg, testViewSpec)

	numRestarts := 6
	// Add some pseudo-random timings
	rnd := rand.New(rand.NewSource(4634653))
	time.Sleep(time.Second)
	for i := 0; i < numRestarts; i++ {
		// wait up to a second
		time.Sleep(time.Millisecond * time.Duration(rnd.Int63n(1000)))
		for ed := range runConsumer(t, cfg) {
			t.Logf("testResume, got block: %d", ed.BlockHeight)
		}
	}
}

func testInvalidUTF8(t *testing.T, cfg *config.VentConfig, tcli rpctransact.TransactClient, inputAddress crypto.Address) {
	create := test.CreateContract(t, tcli, inputAddress)

	// The code point for ó is less than 255 but needs two unicode bytes - it's value expressed as a single byte
	// is in the private use area so is invalid.
	goodString := "Cliente - Doc. identificación"

	// generate events
	name := service.BadStringToHexFunction(goodString)
	description := "Description of TestEvent1"
	test.CallAddEvent(t, tcli, inputAddress, create.Receipt.ContractAddress, name, description)

	// create test db
	_, closeDB := test.NewTestDB(t, cfg)
	defer closeDB()
	resolveSpec(cfg, testViewSpec)

	// Run the consumer with this event - this used to create an error on UPSERT
	runConsumer(t, cfg)

	// Error we used to get before fixing this test case:
	//require.Error(t, err)
	//require.Contains(t, err.Error(), "pq: invalid byte sequence for encoding \"UTF8\": 0xf3 0x6e")
}

func resolveSpec(cfg *config.VentConfig, specFile string) {
	// Resolve relative path to test dir
	_, testFile, _, _ := runtime.Caller(0)
	testDir := path.Join(path.Dir(testFile), "..", "test")

	cfg.SpecFileOrDirs = []string{path.Join(testDir, specFile)}
	cfg.AbiFileOrDirs = []string{path.Join(testDir, "EventsTest.abi")}
	cfg.SpecOpt = sqlsol.BlockTx
}

// Run consumer to listen to events
func runConsumer(t *testing.T, cfg *config.VentConfig) chan types.EventData {
	ch := make(chan types.EventData, 100)
	consumer := service.NewConsumer(cfg, logger, ch)

	projection, err := sqlsol.SpecLoader(cfg.SpecFileOrDirs, cfg.SpecOpt)
	require.NoError(t, err)

	err = consumer.Run(projection, false)
	require.NoError(t, err)
	return consumer.EventsChannel
}
