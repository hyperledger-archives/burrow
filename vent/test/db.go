package test

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

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
func NewTestDB(t *testing.T, cfg *config.VentConfig) (*sqldb.SQLDB, func()) {
	t.Helper()

	if cfg.DBAdapter != types.SQLiteDB {
		if dbURL, ok := syscall.Getenv("DB_URL"); ok {
			t.Logf("Using DB_URL '%s'", dbURL)
			cfg.DBURL = dbURL
		}
	}

	connection := types.SQLConnection{
		DBAdapter: cfg.DBAdapter,
		DBURL:     cfg.DBURL,
		DBSchema:  cfg.DBSchema,

		Log:           logger.NewLogger(""),
		ChainID:       "ID 0123",
		BurrowVersion: "Version 0.0",
	}

	db, err := sqldb.NewSQLDB(connection)
	if err != nil {
		require.NoError(t, err)
	}

	return db, func() {
		if cfg.DBAdapter == types.SQLiteDB {
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

func SqliteVentConfig(grpcAddress string) *config.VentConfig {
	cfg := config.DefaultVentConfig()
	file, err := ioutil.TempFile("", "vent.sqlite")
	if err != nil {
		panic(err)
	}
	err = file.Close()
	if err != nil {
		panic(err)
	}

	cfg.DBURL = file.Name()
	cfg.DBAdapter = types.SQLiteDB
	cfg.GRPCAddr = grpcAddress
	return cfg
}

func PostgresVentConfig(grpcAddress string) *config.VentConfig {
	cfg := config.DefaultVentConfig()
	cfg.DBSchema = fmt.Sprintf("test_%s", randString(10))
	cfg.DBAdapter = types.PostgresDB
	cfg.DBURL = config.DefaultPostgresDBURL
	cfg.GRPCAddr = grpcAddress
	return cfg
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

func randString(n int) string {
	b := make([]rune, n)

	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}
