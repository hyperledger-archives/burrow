// +build integration sqlite

package service_test

import (
	"testing"

	"github.com/hyperledger/burrow/vent/test"
)

func TestSqliteConsumer(t *testing.T) {
	testConsumer(t, test.SqliteVentConfig())
}

func TestSqliteInvalidUTF8(t *testing.T) {
	testInvalidUTF8(t, test.SqliteVentConfig())
}

func TestSqliteDeleteEvent(t *testing.T) {
	testDeleteEvent(t, test.SqliteVentConfig())
}

func TestSqliteResume(t *testing.T) {
	testResume(t, test.SqliteVentConfig())
}
