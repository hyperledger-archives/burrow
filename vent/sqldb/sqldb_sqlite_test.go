// +build integration sqlite

package sqldb_test

import (
	"testing"

	"github.com/hyperledger/burrow/vent/config"
)

func TestSqliteSynchronizeDB(t *testing.T) {
	testSynchronizeDB(t, config.DefaultSqliteFlags())
}

func TestSqliteCleanDB(t *testing.T) {
	testCleanDB(t, config.DefaultSqliteFlags())
}

func TestSqliteSetBlock(t *testing.T) {
	testSetBlock(t, config.DefaultSqliteFlags())
}
