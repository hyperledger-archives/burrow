// sqlite3 is a CGO dependency - we cannot have it on board if we want to use pure Go (e.g. for cross-compiling and other things)
// +build sqlite

package adapters

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/hyperledger/burrow/vent/logger"
	"github.com/hyperledger/burrow/vent/types"
	sqlite3 "github.com/mattn/go-sqlite3"
)

var sqliteDataTypes = map[types.SQLColumnType]string{
	types.SQLColumnTypeBool:      "BOOLEAN",
	types.SQLColumnTypeByteA:     "BLOB",
	types.SQLColumnTypeInt:       "INTEGER",
	types.SQLColumnTypeSerial:    "SERIAL",
	types.SQLColumnTypeText:      "TEXT",
	types.SQLColumnTypeVarchar:   "VARCHAR",
	types.SQLColumnTypeTimeStamp: "TIMESTAMP",
	types.SQLColumnTypeNumeric:   "NUMERIC",
	types.SQLColumnTypeJSON:      "TEXT",
	types.SQLColumnTypeBigInt:    "BIGINT",
}

// SQLiteAdapter implements DBAdapter for SQLiteDB
type SQLiteAdapter struct {
	Log *logger.Logger
}

// NewSQLiteAdapter constructs a new db adapter
func NewSQLiteAdapter(log *logger.Logger) *SQLiteAdapter {
	return &SQLiteAdapter{
		Log: log,
	}
}

// Open connects to a SQLiteQL database, opens it & create default schema if provided
func (adapter *SQLiteAdapter) Open(dbURL string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbURL)
	if err != nil {
		adapter.Log.Info("msg", "Error creating database connection", "err", err)
		return nil, err
	}

	return db, nil
}

// TypeMapping convert generic dataTypes to database dependent dataTypes
func (adapter *SQLiteAdapter) TypeMapping(sqlColumnType types.SQLColumnType) (string, error) {
	if sqlDataType, ok := sqliteDataTypes[sqlColumnType]; ok {
		return sqlDataType, nil
	}

	return "", fmt.Errorf("datatype %v not recognized", sqlColumnType)
}

// SecureColumnName return columns between appropriate security containers
func (adapter *SQLiteAdapter) SecureName(name string) string {
	return Cleanf("[%s]", name)
}

// CreateTableQuery builds query for creating a new table
func (adapter *SQLiteAdapter) CreateTableQuery(tableName string, columns []*types.SQLTableColumn) (string, string) {
	// build query
	columnsDef := ""
	primaryKey := ""
	dictionaryValues := ""
	hasSerial := false

	for i, column := range columns {
		secureColumn := adapter.SecureName(column.Name)
		sqlType, _ := adapter.TypeMapping(column.Type)
		pKey := 0

		if columnsDef != "" {
			columnsDef += ", "
			dictionaryValues += ", "
		}

		if column.Type == types.SQLColumnTypeSerial {
			// SQLITE AUTOINCREMENT LIMITATION
			columnsDef += Cleanf("%s %s", secureColumn, "INTEGER PRIMARY KEY AUTOINCREMENT")
			hasSerial = true
		} else {
			columnsDef += Cleanf("%s %s", secureColumn, sqlType)
		}

		if column.Length > 0 {
			columnsDef += Cleanf("(%v)", column.Length)
		}

		if column.Primary {
			pKey = 1
			columnsDef += " NOT NULL"
			if primaryKey != "" {
				primaryKey += ", "
			}
			primaryKey += secureColumn
		}

		dictionaryValues += Cleanf("('%s','%s',%d,%d,%d,%d)",
			tableName,
			column.Name,
			column.Type,
			column.Length,
			pKey,
			i)
	}

	query := Cleanf("CREATE TABLE %s (%s", adapter.SecureName(tableName), columnsDef)
	if primaryKey != "" {
		if hasSerial {
			// SQLITE AUTOINCREMENT LIMITATION
			query += "," + Cleanf("UNIQUE (%s)", primaryKey)
		} else {
			query += "," + Cleanf("CONSTRAINT %s_pkey PRIMARY KEY (%s)", tableName, primaryKey)
		}
	}
	query += ");"

	dictionaryQuery := Cleanf("INSERT INTO %s (%s,%s,%s,%s,%s,%s) VALUES %s;",
		types.SQLDictionaryTableName,
		types.SQLColumnLabelTableName, types.SQLColumnLabelColumnName,
		types.SQLColumnLabelColumnType, types.SQLColumnLabelColumnLength,
		types.SQLColumnLabelPrimaryKey, types.SQLColumnLabelColumnOrder,
		dictionaryValues)

	return query, dictionaryQuery
}

// LastBlockIDQuery returns a query for last inserted blockId in log table
func (adapter *SQLiteAdapter) LastBlockIDQuery() string {
	query := `
		WITH ll AS (
			SELECT MAX(%s) AS %s FROM %s WHERE %s = $1
		)
		SELECT COALESCE(%s, '0') AS %s
			FROM ll LEFT OUTER JOIN %s log ON (ll.%s = log.%s);`

	return Cleanf(query,
		types.SQLColumnLabelId,               // max
		types.SQLColumnLabelId,               // as
		types.SQLLogTableName,                // from
		types.SQLColumnLabelChainID, chainid, // where
		types.SQLColumnLabelHeight,                     // coalesce
		types.SQLColumnLabelHeight,                     // as
		types.SQLLogTableName,                          // from
		types.SQLColumnLabelId, types.SQLColumnLabelId) // on
}

// FindTableQuery returns a query that checks if a table exists
func (adapter *SQLiteAdapter) FindTableQuery() string {
	query := "SELECT COUNT(*) found FROM %s WHERE %s = $1;"

	return Cleanf(query,
		types.SQLDictionaryTableName,  // from
		types.SQLColumnLabelTableName) // where
}

// TableDefinitionQuery returns a query with table structure
func (adapter *SQLiteAdapter) TableDefinitionQuery() string {
	query := `
		SELECT
			%s,%s,%s,%s
		FROM
			%s
		WHERE
			%s = $1
		ORDER BY
			%s;`

	return Cleanf(query,
		types.SQLColumnLabelColumnName, types.SQLColumnLabelColumnType, // select
		types.SQLColumnLabelColumnLength, types.SQLColumnLabelPrimaryKey, // select
		types.SQLDictionaryTableName,    // from
		types.SQLColumnLabelTableName,   // where
		types.SQLColumnLabelColumnOrder) // order by
}

// AlterColumnQuery returns a query for adding a new column to a table
func (adapter *SQLiteAdapter) AlterColumnQuery(tableName, columnName string, sqlColumnType types.SQLColumnType, length, order int) (string, string) {
	sqlType, _ := adapter.TypeMapping(sqlColumnType)
	if length > 0 {
		sqlType = Cleanf("%s(%d)", sqlType, length)
	}

	query := Cleanf("ALTER TABLE %s ADD COLUMN %s %s;",
		adapter.SecureName(tableName),
		adapter.SecureName(columnName),
		sqlType)

	dictionaryQuery := Cleanf(`
		INSERT INTO %s (%s,%s,%s,%s,%s,%s)
		VALUES ('%s','%s',%d,%d,%d,%d);`,

		types.SQLDictionaryTableName,

		types.SQLColumnLabelTableName, types.SQLColumnLabelColumnName,
		types.SQLColumnLabelColumnType, types.SQLColumnLabelColumnLength,
		types.SQLColumnLabelPrimaryKey, types.SQLColumnLabelColumnOrder,

		tableName, columnName, sqlColumnType, length, 0, order)

	return query, dictionaryQuery
}

// SelectRowQuery returns a query for selecting row values
func (adapter *SQLiteAdapter) SelectRowQuery(tableName, fields, indexValue string) string {
	return Cleanf("SELECT %s FROM %s WHERE %s = '%s';", fields, adapter.SecureName(tableName), types.SQLColumnLabelHeight, indexValue)
}

// SelectLogQuery returns a query for selecting all tables involved in a block trn
func (adapter *SQLiteAdapter) SelectLogQuery() string {
	query := `
		SELECT DISTINCT %s,%s FROM %s l WHERE %s = $1 AMD %s = $2;`

	return Cleanf(query,
		types.SQLColumnLabelTableName, types.SQLColumnLabelEventName, // select
		types.SQLLogTableName, // from
		types.SQLColumnLabelChainID,
		types.SQLColumnLabelHeight) // where
}

// InsertLogQuery returns a query to insert a row in log table
func (adapter *SQLiteAdapter) InsertLogQuery() string {
	query := `
		INSERT INTO %s ($s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)
		VALUES (CURRENT_TIMESTAMP, $1, $2, $3, $4, $5, $6 ,$7, $8, $9, $10);`

	return Cleanf(query,
		types.SQLLogTableName, // insert
		//fields
		types.SQLColumnLabelChainID,
		types.SQLColumnLabelTimeStamp, types.SQLColumnLabelTableName, types.SQLColumnLabelEventName, types.SQLColumnLabelEventFilter,
		types.SQLColumnLabelHeight, types.SQLColumnLabelTxHash, types.SQLColumnLabelAction, types.SQLColumnLabelDataRow,
		types.SQLColumnLabelSqlStmt, types.SQLColumnLabelSqlValues)
}

// ErrorEquals verify if an error is of a given SQL type
func (adapter *SQLiteAdapter) ErrorEquals(err error, sqlErrorType types.SQLErrorType) bool {
	if err, ok := err.(sqlite3.Error); ok {
		errDescription := err.Error()

		switch sqlErrorType {
		case types.SQLErrorTypeGeneric:
			return true
		case types.SQLErrorTypeDuplicatedColumn:
			return err.Code == 1 && strings.Contains(errDescription, "duplicate column")
		case types.SQLErrorTypeDuplicatedTable:
			return err.Code == 1 && strings.Contains(errDescription, "table") && strings.Contains(errDescription, "already exists")
		case types.SQLErrorTypeUndefinedTable:
			return err.Code == 1 && strings.Contains(errDescription, "no such table")
		case types.SQLErrorTypeUndefinedColumn:
			return err.Code == 1 && strings.Contains(errDescription, "table") && strings.Contains(errDescription, "has no column named")
		case types.SQLErrorTypeInvalidType:
			// NOT SUPPORTED
			return false
		}
	}

	return false
}

func (adapter *SQLiteAdapter) UpsertQuery(table *types.SQLTable, row types.EventDataRow) (types.UpsertDeleteQuery, interface{}, error) {
	pointers := make([]interface{}, 0)
	columns := ""
	insValues := ""
	updValues := ""
	pkColumns := ""
	values := ""
	var txHash interface{} = nil

	i := 0

	// for each column in table
	for _, column := range table.Columns {
		secureColumn := adapter.SecureName(column.Name)

		i++

		// INSERT INTO TABLE (*columns).........
		if columns != "" {
			columns += ", "
			insValues += ", "
			values += ", "
		}
		columns += secureColumn
		insValues += "$" + Cleanf("%d", i)

		//find data for column
		if value, ok := row.RowData[column.Name]; ok {
			//load hash value
			if column.Name == types.SQLColumnLabelTxHash {
				txHash = value
			}

			// column found (not null)
			// load values
			pointers = append(pointers, &value)
			values += fmt.Sprint(value)

			if !column.Primary {
				// column is no PK
				// add to update list
				// INSERT........... ON CONFLICT......DO UPDATE (*updValues)
				if updValues != "" {
					updValues += ", "
				}
				updValues += secureColumn + " = $" + Cleanf("%d", i)
			}
		} else if column.Primary {
			// column NOT found (is null) and is PK
			return types.UpsertDeleteQuery{}, nil, fmt.Errorf("error null primary key for column %s", secureColumn)
		} else {
			// column NOT found (is null) and is NOT PK
			pointers = append(pointers, nil)
			values += "null"
		}

		if column.Primary {
			// ON CONFLICT (....values....)
			if pkColumns != "" {
				pkColumns += ", "
			}
			pkColumns += secureColumn
		}

	}

	query := Cleanf("INSERT INTO %s (%s) VALUES (%s) ", adapter.SecureName(table.Name), columns, insValues)

	if pkColumns != "" {
		if updValues != "" {
			query += Cleanf("ON CONFLICT (%s) DO UPDATE SET %s", pkColumns, updValues)
		} else {
			query += Cleanf("ON CONFLICT (%s) DO NOTHING", pkColumns)
		}
	}
	query += ";"

	return types.UpsertDeleteQuery{Query: query, Values: values, Pointers: pointers}, txHash, nil
}

func (adapter *SQLiteAdapter) DeleteQuery(table *types.SQLTable, row types.EventDataRow) (types.UpsertDeleteQuery, error) {

	pointers := make([]interface{}, 0)
	columns := ""
	values := ""
	i := 0

	// for each column in table
	for _, column := range table.Columns {

		//only PK for delete
		if column.Primary {
			i++

			secureColumn := adapter.SecureName(column.Name)

			// WHERE ..........
			if columns != "" {
				columns += " AND "
				values += ", "
			}

			columns += Cleanf("%s = $%d", secureColumn, i)

			//find data for column
			if value, ok := row.RowData[column.Name]; ok {
				// column found (not null)
				// load values
				pointers = append(pointers, &value)
				values += fmt.Sprint(value)

			} else {
				// column NOT found (is null) and is PK
				return types.UpsertDeleteQuery{}, fmt.Errorf("error null primary key for column %s", secureColumn)
			}
		}
	}

	if columns == "" {
		return types.UpsertDeleteQuery{}, fmt.Errorf("error primary key not found for deletion")
	}

	query := Cleanf("DELETE FROM %s WHERE %s;", adapter.SecureName(table.Name), columns)

	return types.UpsertDeleteQuery{Query: query, Values: values, Pointers: pointers}, nil
}

func (adapter *SQLiteAdapter) RestoreDBQuery() string {

	query := Cleanf("SELECT %s, %s, %s, %s FROM %s",
		types.SQLColumnLabelTableName, types.SQLColumnLabelAction, types.SQLColumnLabelSqlStmt, types.SQLColumnLabelSqlValues,
		types.SQLLogTableName)

	query += " WHERE strftime('%Y-%m-%d %H:%M:%S',"

	query += Cleanf("%s)<=$1 ORDER BY %s;",
		types.SQLColumnLabelTimeStamp, types.SQLColumnLabelId)

	return query

}

func (adapter *SQLiteAdapter) CleanDBQueries() types.SQLCleanDBQuery {
	// Chain info
	selectChainIDQry := Cleanf(`
		SELECT 
		COUNT(*) REGISTERS,
		COALESCE(MAX(%s),'') CHAINID,
		COALESCE(MAX(%s),'') BVERSION 
		FROM %s;`,
		types.SQLColumnLabelChainID, types.SQLColumnLabelBurrowVer,
		types.SQLChainInfoTableName)

	deleteChainIDQry := Cleanf(`
		DELETE FROM %s;`,
		types.SQLChainInfoTableName)

	insertChainIDQry := Cleanf(`
		INSERT INTO %s (%s,%s) VALUES($1,$2)`,
		types.SQLChainInfoTableName,
		types.SQLColumnLabelChainID, types.SQLColumnLabelBurrowVer)

	// Dictionary
	selectDictionaryQry := Cleanf(`
		SELECT DISTINCT %s 
		FROM %s 
 		WHERE %s
		NOT IN ('%s','%s','%s');`,
		types.SQLColumnLabelTableName,
		types.SQLDictionaryTableName,
		types.SQLColumnLabelTableName,
		types.SQLLogTableName, types.SQLDictionaryTableName, types.SQLChainInfoTableName)

	deleteDictionaryQry := Cleanf(`
		DELETE FROM %s 
		WHERE %s 
		NOT IN ('%s','%s','%s');`,
		types.SQLDictionaryTableName,
		types.SQLColumnLabelTableName,
		types.SQLLogTableName, types.SQLDictionaryTableName, types.SQLChainInfoTableName)

	// log
	deleteLogQry := Cleanf(`
		DELETE FROM %s;`,
		types.SQLLogTableName)

	return types.SQLCleanDBQuery{
		SelectChainIDQry:    selectChainIDQry,
		DeleteChainIDQry:    deleteChainIDQry,
		InsertChainIDQry:    insertChainIDQry,
		SelectDictionaryQry: selectDictionaryQry,
		DeleteDictionaryQry: deleteDictionaryQry,
		DeleteLogQry:        deleteLogQry,
	}
}

func (adapter *SQLiteAdapter) DropTableQuery(tableName string) string {
	// SQLite does not support DROP TABLE CASCADE so this will fail if there are dependent objects
	return Cleanf(`DROP TABLE %s;`, adapter.SecureName(tableName))
}
