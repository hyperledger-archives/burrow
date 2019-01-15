// +build integration

package sqldb_test

import (
	"testing"

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
