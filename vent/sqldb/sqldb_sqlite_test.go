// +build integration sqlite

package sqldb_test

import (
	"testing"

	"github.com/hyperledger/burrow/vent/test"
)

func TestSqliteSynchronizeDB(t *testing.T) {
	testSynchronizeDB(t, test.SqliteFlags())
}

func TestSqliteCleanDB(t *testing.T) {
	testCleanDB(t, test.SqliteFlags())
}

func TestSqliteSetBlock(t *testing.T) {
	testSetBlock(t, test.SqliteFlags())
}
