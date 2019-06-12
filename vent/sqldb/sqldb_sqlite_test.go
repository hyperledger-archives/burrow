// +build integration sqlite

package sqldb_test

import (
	"testing"

	"github.com/hyperledger/burrow/vent/test"
)

func TestSqliteSynchronizeDB(t *testing.T) {
	testSynchronizeDB(t, test.SqliteVentConfig(""))
}

func TestSqliteCleanDB(t *testing.T) {
	testCleanDB(t, test.SqliteVentConfig(""))
}

func TestSqliteSetBlock(t *testing.T) {
	testSetBlock(t, test.SqliteVentConfig(""))
}

func TestSqliteRestore(t *testing.T) {
	testRestore(t, test.SqliteVentConfig(""))
}
