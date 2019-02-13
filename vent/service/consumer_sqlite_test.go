// +build integration sqlite

package service_test

import (
	"testing"

	"github.com/hyperledger/burrow/vent/test"
)

func TestSqliteConsumer(t *testing.T) {
	testConsumer(t, test.SqliteFlags())
}

func TestSqliteInvalidUTF8(t *testing.T) {
	testInvalidUTF8(t, test.SqliteFlags())
}

func TestSqliteDeleteEvent(t *testing.T) {
	testDeleteEvent(t, test.SqliteFlags())
}

func TestSqliteResume(t *testing.T) {
	testResume(t, test.SqliteFlags())
}
