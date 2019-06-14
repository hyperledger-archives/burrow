// sqlite3 is a CGO dependency - we cannot have it on board if we want to use pure Go (e.g. for cross-compiling and other things)
// +build !sqlite

package adapters

import (
	"fmt"

	"github.com/hyperledger/burrow/vent/logger"
	"github.com/hyperledger/burrow/vent/types"
	"github.com/jmoiron/sqlx"
)

// This is a no-op version of SQLiteAdapter
type SQLiteAdapter struct {
	Log *logger.Logger
}

var _ DBAdapter = &SQLiteAdapter{}

func NewSQLiteAdapter(names types.SQLNames, log *logger.Logger) *SQLiteAdapter {
	panic(fmt.Errorf("vent has been built without sqlite support. To use the sqlite DBAdapter build with the 'sqlite' build tag enabled"))
}

func (*SQLiteAdapter) Open(dbURL string) (*sqlx.DB, error) {
	panic("implement me")
}

func (*SQLiteAdapter) TypeMapping(sqlColumnType types.SQLColumnType) (string, error) {
	panic("implement me")
}

func (*SQLiteAdapter) ErrorEquals(err error, sqlErrorType types.SQLErrorType) bool {
	panic("implement me")
}

func (*SQLiteAdapter) SecureName(name string) string {
	panic("implement me")
}

func (*SQLiteAdapter) CreateTableQuery(tableName string, columns []*types.SQLTableColumn) (string, string) {
	panic("implement me")
}

func (*SQLiteAdapter) FindTableQuery() string {
	panic("implement me")
}

func (*SQLiteAdapter) TableDefinitionQuery() string {
	panic("implement me")
}

func (*SQLiteAdapter) AlterColumnQuery(tableName, columnName string, sqlColumnType types.SQLColumnType, length, order int) (string, string) {
	panic("implement me")
}

func (*SQLiteAdapter) SelectRowQuery(tableName, fields, indexValue string) string {
	panic("implement me")
}

func (*SQLiteAdapter) SelectLogQuery() string {
	panic("implement me")
}

func (*SQLiteAdapter) InsertLogQuery() string {
	panic("implement me")
}

func (*SQLiteAdapter) UpsertQuery(table *types.SQLTable, row types.EventDataRow) (types.UpsertDeleteQuery, interface{}, error) {
	panic("implement me")
}

func (*SQLiteAdapter) DeleteQuery(table *types.SQLTable, row types.EventDataRow) (types.UpsertDeleteQuery, error) {
	panic("implement me")
}

func (*SQLiteAdapter) RestoreDBQuery() string {
	panic("implement me")
}

func (*SQLiteAdapter) CleanDBQueries() types.SQLCleanDBQuery {
	panic("implement me")
}

func (*SQLiteAdapter) DropTableQuery(tableName string) string {
	panic("implement me")
}

func (*SQLiteAdapter) SchemaName(tableName string) string {
	panic("implement me")
}
