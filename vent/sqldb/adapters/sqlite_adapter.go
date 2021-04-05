// sqlite3 is a CGO dependency - we cannot have it on board if we want to use pure Go (e.g. for cross-compiling and other things)
// +build sqlite

package adapters

import (
	"fmt"
	"strings"

	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/vent/types"
	"github.com/jmoiron/sqlx"
	sqlite3 "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
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
	types.SQLNames
	Log *logging.Logger
}

var _ DBAdapter = &SQLiteAdapter{}

// NewSQLiteAdapter constructs a new db adapter
func NewSQLiteAdapter(sqlNames types.SQLNames, log *logging.Logger) *SQLiteAdapter {
	return &SQLiteAdapter{
		SQLNames: sqlNames,
		Log:      log,
	}
}

func (sla *SQLiteAdapter) Open(dbURL string) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite3", dbURL)
	if err != nil {
		sla.Log.InfoMsg("Error creating database connection", "err", err)
		return nil, err
	}
	return db, nil
}

// TypeMapping convert generic dataTypes to database dependent dataTypes
func (sla *SQLiteAdapter) TypeMapping(sqlColumnType types.SQLColumnType) (string, error) {
	if sqlDataType, ok := sqliteDataTypes[sqlColumnType]; ok {
		return sqlDataType, nil
	}

	return "", fmt.Errorf("datatype %v not recognized", sqlColumnType)
}

// SecureColumnName return columns between appropriate security containers
func (sla *SQLiteAdapter) SecureName(name string) string {
	return Cleanf("[%s]", name)
}

// CreateTableQuery builds query for creating a new table
func (sla *SQLiteAdapter) CreateTableQuery(tableName string, columns []*types.SQLTableColumn) (string, string) {
	// build query
	columnsDef := ""
	primaryKey := ""
	dictionaryValues := ""
	hasSerial := false

	for i, column := range columns {
		secureColumn := sla.SecureName(column.Name)
		sqlType, _ := sla.TypeMapping(column.Type)
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

	query := Cleanf("CREATE TABLE %s (%s", sla.SecureName(tableName), columnsDef)
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
		sla.Tables.Dictionary,
		sla.Columns.TableName, sla.Columns.ColumnName,
		sla.Columns.ColumnType, sla.Columns.ColumnLength,
		sla.Columns.PrimaryKey, sla.Columns.ColumnOrder,
		dictionaryValues)

	return query, dictionaryQuery
}

// FindTableQuery returns a query that checks if a table exists
func (sla *SQLiteAdapter) FindTableQuery() string {
	query := "SELECT COUNT(*) found FROM %s WHERE %s = $1;"

	return Cleanf(query,
		sla.Tables.Dictionary, // from
		sla.Columns.TableName) // where
}

// TableDefinitionQuery returns a query with table structure
func (sla *SQLiteAdapter) TableDefinitionQuery() string {
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
		sla.Columns.ColumnName, sla.Columns.ColumnType, // select
		sla.Columns.ColumnLength, sla.Columns.PrimaryKey, // select
		sla.Tables.Dictionary,   // from
		sla.Columns.TableName,   // where
		sla.Columns.ColumnOrder) // order by
}

// AlterColumnQuery returns a query for adding a new column to a table
func (sla *SQLiteAdapter) AlterColumnQuery(tableName, columnName string, sqlColumnType types.SQLColumnType, length, order int) (string, string) {
	sqlType, _ := sla.TypeMapping(sqlColumnType)
	if length > 0 {
		sqlType = Cleanf("%s(%d)", sqlType, length)
	}

	query := Cleanf("ALTER TABLE %s ADD COLUMN %s %s;",
		sla.SecureName(tableName),
		sla.SecureName(columnName),
		sqlType)

	dictionaryQuery := Cleanf(`
		INSERT INTO %s (%s,%s,%s,%s,%s,%s)
		VALUES ('%s','%s',%d,%d,%d,%d);`,

		sla.Tables.Dictionary,

		sla.Columns.TableName, sla.Columns.ColumnName,
		sla.Columns.ColumnType, sla.Columns.ColumnLength,
		sla.Columns.PrimaryKey, sla.Columns.ColumnOrder,

		tableName, columnName, sqlColumnType, length, 0, order)

	return query, dictionaryQuery
}

// SelectRowQuery returns a query for selecting row values
func (sla *SQLiteAdapter) SelectRowQuery(tableName, fields, indexValue string) string {
	return Cleanf("SELECT %s FROM %s WHERE %s = '%s';", fields, sla.SecureName(tableName), sla.Columns.Height, indexValue)
}

// SelectLogQuery returns a query for selecting all tables involved in a block trn
func (sla *SQLiteAdapter) SelectLogQuery() string {
	query := `
		SELECT DISTINCT %s,%s FROM %s l WHERE %s = $1 AND %s = $2;`

	return Cleanf(query,
		sla.Columns.TableName, sla.Columns.EventName, // select
		sla.Tables.Log,     // from
		sla.Columns.Height, // where
		sla.Columns.ChainID)
}

// InsertLogQuery returns a query to insert a row in log table
func (sla *SQLiteAdapter) InsertLogQuery() string {
	query := `
		INSERT INTO %s (%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)
		VALUES (CURRENT_TIMESTAMP, $1, $2, $3, $4, $5, $6 ,$7, $8, $9, $10);`

	return Cleanf(query,
		sla.Tables.Log, // insert
		//fields
		sla.Columns.TimeStamp,
		sla.Columns.ChainID,
		sla.Columns.TableName, sla.Columns.EventName, sla.Columns.EventFilter,
		sla.Columns.Height, sla.Columns.TxHash, sla.Columns.Action, sla.Columns.DataRow,
		sla.Columns.SqlStmt, sla.Columns.SqlValues)
}

// ErrorEquals verify if an error is of a given SQL type
func (sla *SQLiteAdapter) ErrorEquals(err error, sqlErrorType types.SQLErrorType) bool {
	slErr := new(sqlite3.Error)
	if errors.As(err, slErr) {
		errDescription := err.Error()

		switch sqlErrorType {
		case types.SQLErrorTypeGeneric:
			return true
		case types.SQLErrorTypeDuplicatedColumn:
			return slErr.Code == 1 && strings.Contains(errDescription, "duplicate column")
		case types.SQLErrorTypeDuplicatedTable:
			return slErr.Code == 1 && strings.Contains(errDescription, "table") && strings.Contains(errDescription, "already exists")
		case types.SQLErrorTypeUndefinedTable:
			return slErr.Code == 1 && strings.Contains(errDescription, "no such table")
		case types.SQLErrorTypeUndefinedColumn:
			return slErr.Code == 1 && strings.Contains(errDescription, "table") && strings.Contains(errDescription, "has no column named")
		case types.SQLErrorTypeInvalidType:
			// NOT SUPPORTED
			return false
		}
	}

	return false
}

func (sla *SQLiteAdapter) UpsertQuery(table *types.SQLTable, row types.EventDataRow) (types.UpsertDeleteQuery, interface{}, error) {
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
		secureColumn := sla.SecureName(column.Name)

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
			if column.Name == sla.Columns.TxHash {
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

	query := Cleanf("INSERT INTO %s (%s) VALUES (%s) ", sla.SecureName(table.Name), columns, insValues)

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

func (sla *SQLiteAdapter) DeleteQuery(table *types.SQLTable, row types.EventDataRow) (types.UpsertDeleteQuery, error) {

	pointers := make([]interface{}, 0)
	columns := ""
	values := ""
	i := 0

	// for each column in table
	for _, column := range table.Columns {

		//only PK for delete
		if column.Primary {
			i++

			secureColumn := sla.SecureName(column.Name)

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

	query := Cleanf("DELETE FROM %s WHERE %s;", sla.SecureName(table.Name), columns)

	return types.UpsertDeleteQuery{Query: query, Values: values, Pointers: pointers}, nil
}

func (sla *SQLiteAdapter) RestoreDBQuery() string {

	query := Cleanf("SELECT %s, %s, %s, %s, %s FROM %s",
		sla.Columns.Id, sla.Columns.TableName, sla.Columns.Action, // select
		sla.Columns.SqlStmt, sla.Columns.SqlValues, // select
		sla.Tables.Log)

	query += " WHERE strftime('%Y-%m-%d %H:%M:%S',"

	query += Cleanf("%s) AND %s != '%s' AND  %s != '%s' <=$1 ORDER BY %s;",
		sla.Columns.TimeStamp,
		sla.Columns.TableName, sla.Tables.Block, // where not _vent_block
		sla.Columns.TableName, sla.Tables.Tx, // where not _vent_tx
		sla.Columns.Id)

	return query

}

func (sla *SQLiteAdapter) CleanDBQueries() types.SQLCleanDBQuery {
	// Chain info
	selectChainIDQry := Cleanf(`
		SELECT 
		COUNT(*) REGISTERS,
		COALESCE(MAX(%s),'') CHAINID,
		COALESCE(MAX(%s),'') BVERSION 
		FROM %s;`,
		sla.Columns.ChainID, sla.Columns.BurrowVersion,
		sla.Tables.ChainInfo)

	deleteChainIDQry := Cleanf(`
		DELETE FROM %s;`,
		sla.Tables.ChainInfo)

	insertChainIDQry := Cleanf(`
		INSERT INTO %s (%s,%s,%s) VALUES($1,$2,$3)`,
		sla.Tables.ChainInfo,
		sla.Columns.ChainID, sla.Columns.BurrowVersion, sla.Columns.Height)

	// Dictionary
	selectDictionaryQry := Cleanf(`
		SELECT DISTINCT %s 
		FROM %s 
 		WHERE %s
		NOT IN ('%s','%s','%s');`,
		sla.Columns.TableName,
		sla.Tables.Dictionary,
		sla.Columns.TableName,
		sla.Tables.Log, sla.Tables.Dictionary, sla.Tables.ChainInfo)

	deleteDictionaryQry := Cleanf(`
		DELETE FROM %s 
		WHERE %s 
		NOT IN ('%s','%s','%s');`,
		sla.Tables.Dictionary,
		sla.Columns.TableName,
		sla.Tables.Log, sla.Tables.Dictionary, sla.Tables.ChainInfo)

	// log
	deleteLogQry := Cleanf(`
		DELETE FROM %s;`,
		sla.Tables.Log)

	return types.SQLCleanDBQuery{
		SelectChainIDQry:    selectChainIDQry,
		DeleteChainIDQry:    deleteChainIDQry,
		InsertChainIDQry:    insertChainIDQry,
		SelectDictionaryQry: selectDictionaryQry,
		DeleteDictionaryQry: deleteDictionaryQry,
		DeleteLogQry:        deleteLogQry,
	}
}

func (sla *SQLiteAdapter) DropTableQuery(tableName string) string {
	// SQLite does not support DROP TABLE CASCADE so this will fail if there are dependent objects
	return Cleanf(`DROP TABLE IF EXISTS %s;`, sla.SecureName(tableName))
}

func (sla *SQLiteAdapter) SchemaName(tableName string) string {
	return secureName(tableName)
}
