package test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"os"

	"github.com/hyperledger/burrow/vent/config"
	"github.com/hyperledger/burrow/vent/logger"
	"github.com/hyperledger/burrow/vent/sqldb"
	"github.com/hyperledger/burrow/vent/types"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyz")

func init() {
	rand.Seed(time.Now().UnixNano())
}

// NewTestDB creates a database connection for testing
func NewTestDB(t *testing.T, dbAdapter string) (*sqldb.SQLDB, func()) {
	t.Helper()

	cfg := config.DefaultFlags()

	connection := types.SQLConnection{
		DBAdapter:     cfg.DBAdapter,
		DBURL:         cfg.DBURL,
		Log:           logger.NewLogger("debug"),
		ChainID:       "ID 0123",
		BurrowVersion: "Version 0.0",
	}

	switch dbAdapter {
	case types.PostgresDB:
		connection.DBSchema = fmt.Sprintf("test_%s", randString(10))

	case types.SQLiteDB:
		connection.DBAdapter = dbAdapter
		connection.DBURL = fmt.Sprintf("./test_%s.sqlite", randString(10))

	default:
		t.Fatal("invalid database adapter")
	}

	db, err := sqldb.NewSQLDB(connection)
	if err != nil {
		t.Fatal(err.Error())
	}

	return db, func() {
		if dbAdapter == types.SQLiteDB {
			db.Close()
			os.Remove(connection.DBURL)
			os.Remove(connection.DBURL + "-shm")
			os.Remove(connection.DBURL + "-wal")
		} else {
			destroySchema(db, connection.DBSchema)
			db.Close()
		}
	}
}

func randString(n int) string {
	b := make([]rune, n)

	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

func destroySchema(db *sqldb.SQLDB, dbSchema string) error {
	db.Log.Info("msg", "Dropping schema")
	query := fmt.Sprintf("DROP SCHEMA %s CASCADE;", dbSchema)

	db.Log.Info("msg", "Drop schema", "query", query)

	if _, err := db.DB.Exec(query); err != nil {
		db.Log.Info("msg", "Error dropping schema", "err", err)
		return err
	}

	return nil
}
