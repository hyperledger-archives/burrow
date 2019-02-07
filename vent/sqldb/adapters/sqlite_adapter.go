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
	return fmt.Sprintf("[%s]", name)
}

// CreateTableQuery builds query for creating a new table
func (adapter *SQLiteAdapter) CreateTableQuery(tableName string, columns []types.SQLTableColumn) (string, string) {
	// build query
	columnsDef := ""
	primaryKey := ""
	dictionaryValues := ""
	hasSerial := false

	for i, tableColumn := range columns {
		secureColumn := adapter.SecureName(tableColumn.Name)
		sqlType, _ := adapter.TypeMapping(tableColumn.Type)
		pKey := 0

		if columnsDef != "" {
			columnsDef += ", "
			dictionaryValues += ", "
		}

		if tableColumn.Type == types.SQLColumnTypeSerial {
			// SQLITE AUTOINCREMENT LIMITATION
			columnsDef += fmt.Sprintf("%s %s", secureColumn, "INTEGER PRIMARY KEY AUTOINCREMENT")
			hasSerial = true
		} else {
			columnsDef += fmt.Sprintf("%s %s", secureColumn, sqlType)
		}

		if tableColumn.Length > 0 {
			columnsDef += fmt.Sprintf("(%v)", tableColumn.Length)
		}

		if tableColumn.Primary {
			pKey = 1
			columnsDef += " NOT NULL"
			if primaryKey != "" {
				primaryKey += ", "
			}
			primaryKey += secureColumn
		}

		dictionaryValues += fmt.Sprintf("('%s','%s',%d,%d,%d,%d)",
			tableName,
			tableColumn.Name,
			tableColumn.Type,
			tableColumn.Length,
			pKey,
			i)
	}

	query := fmt.Sprintf("CREATE TABLE %s (%s", tableName, columnsDef)
	if primaryKey != "" {
		if hasSerial {
			// SQLITE AUTOINCREMENT LIMITATION
			query += "," + fmt.Sprintf("UNIQUE (%s)", primaryKey)
		} else {
			query += "," + fmt.Sprintf("CONSTRAINT %s_pkey PRIMARY KEY (%s)", tableName, primaryKey)
		}
	}
	query += ");"

	dictionaryQuery := fmt.Sprintf("INSERT INTO %s (%s,%s,%s,%s,%s,%s) VALUES %s;",
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
			SELECT MAX(%s) AS %s FROM %s
		)
		SELECT COALESCE(%s, '0') AS %s
			FROM ll LEFT OUTER JOIN %s log ON (ll.%s = log.%s);`

	return fmt.Sprintf(query,
		types.SQLColumnLabelId,                         // max
		types.SQLColumnLabelId,                         // as
		types.SQLLogTableName,                          // from
		types.SQLColumnLabelHeight,                     // coalesce
		types.SQLColumnLabelHeight,                     // as
		types.SQLLogTableName,                          // from
		types.SQLColumnLabelId, types.SQLColumnLabelId) // on
}

// FindTableQuery returns a query that checks if a table exists
func (adapter *SQLiteAdapter) FindTableQuery() string {
	query := "SELECT COUNT(*) found FROM %s WHERE %s = $1;"

	return fmt.Sprintf(query,
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

	return fmt.Sprintf(query,
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
		sqlType = fmt.Sprintf("%s(%d)", sqlType, length)
	}

	query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s;",
		tableName,
		adapter.SecureName(columnName),
		sqlType)

	dictionaryQuery := fmt.Sprintf(`
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
	return fmt.Sprintf("SELECT %s FROM %s WHERE %s = '%s';", fields, tableName, types.SQLColumnLabelHeight, indexValue)
}

// SelectLogQuery returns a query for selecting all tables involved in a block trn
func (adapter *SQLiteAdapter) SelectLogQuery() string {
	query := `
		SELECT DISTINCT %s,%s FROM %s l WHERE %s = $1;`

	return fmt.Sprintf(query,
		types.SQLColumnLabelTableName, types.SQLColumnLabelEventName, // select
		types.SQLLogTableName,      // from
		types.SQLColumnLabelHeight) // where
}

// InsertLogQuery returns a query to insert a row in log table
func (adapter *SQLiteAdapter) InsertLogQuery() string {
	query := `
		INSERT INTO %s (%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)
		VALUES (CURRENT_TIMESTAMP, $1, $2, $3, $4, $5, $6 ,$7, $8, $9);`

	return fmt.Sprintf(query,
		types.SQLLogTableName, // insert
		//fields
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

func (adapter *SQLiteAdapter) UpsertQuery(table types.SQLTable, row types.EventDataRow) (types.UpsertDeleteQuery, interface{}, error) {

	pointers := make([]interface{}, 0)
	columns := ""
	insValues := ""
	updValues := ""
	pkColumns := ""
	values := ""
	var txHash interface{} = nil

	i := 0

	// for each column in table
	for _, tableColumn := range table.Columns {
		secureColumn := adapter.SecureName(tableColumn.Name)

		i++

		// INSERT INTO TABLE (*columns).........
		if columns != "" {
			columns += ", "
			insValues += ", "
			values += ", "
		}
		columns += secureColumn
		insValues += "$" + fmt.Sprintf("%d", i)

		//find data for column
		if value, ok := row.RowData[tableColumn.Name]; ok {
			//load hash value
			if tableColumn.Name == types.SQLColumnLabelTxHash {
				txHash = value
			}

			// column found (not null)
			// load values
			pointers = append(pointers, &value)
			values += fmt.Sprint(value)

			if !tableColumn.Primary {
				// column is no PK
				// add to update list
				// INSERT........... ON CONFLICT......DO UPDATE (*updValues)
				if updValues != "" {
					updValues += ", "
				}
				updValues += secureColumn + " = $" + fmt.Sprintf("%d", i)
			}
		} else if tableColumn.Primary {
			// column NOT found (is null) and is PK
			return types.UpsertDeleteQuery{}, nil, fmt.Errorf("error null primary key for column %s", secureColumn)
		} else {
			// column NOT found (is null) and is NOT PK
			pointers = append(pointers, nil)
			values += "null"
		}

		if tableColumn.Primary {
			// ON CONFLICT (....values....)
			if pkColumns != "" {
				pkColumns += ", "
			}
			pkColumns += secureColumn
		}

	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ", table.Name, columns, insValues)

	if pkColumns != "" {
		if updValues != "" {
			query += fmt.Sprintf("ON CONFLICT (%s) DO UPDATE SET %s", pkColumns, updValues)
		} else {
			query += fmt.Sprintf("ON CONFLICT (%s) DO NOTHING", pkColumns)
		}
	}
	query += ";"

	return types.UpsertDeleteQuery{Query: query, Values: values, Pointers: pointers}, txHash, nil
}

func (adapter *SQLiteAdapter) DeleteQuery(table types.SQLTable, row types.EventDataRow) (types.UpsertDeleteQuery, error) {

	pointers := make([]interface{}, 0)
	columns := ""
	values := ""
	i := 0

	// for each column in table
	for _, tableColumn := range table.Columns {

		//only PK for delete
		if tableColumn.Primary {
			i++

			secureColumn := adapter.SecureName(tableColumn.Name)

			// WHERE ..........
			if columns != "" {
				columns += "AND "
				values += ", "
			}

			columns += fmt.Sprintf("%s = $%d", secureColumn, i)

			//find data for column
			if value, ok := row.RowData[tableColumn.Name]; ok {
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

	query := fmt.Sprintf("DELETE FROM %s WHERE %s;", table.Name, columns)

	return types.UpsertDeleteQuery{Query: query, Values: values, Pointers: pointers}, nil
}

func (adapter *SQLiteAdapter) RestoreDBQuery() string {

	query := fmt.Sprintf("SELECT %s, %s, %s, %s FROM %s",
		types.SQLColumnLabelTableName, types.SQLColumnLabelAction, types.SQLColumnLabelSqlStmt, types.SQLColumnLabelSqlValues,
		types.SQLLogTableName)

	query += " WHERE strftime('%Y-%m-%d %H:%M:%S',"

	query += fmt.Sprintf("%s)<=$1 ORDER BY %s;",
		types.SQLColumnLabelTimeStamp, types.SQLColumnLabelId)

	return query

}

func (adapter *SQLiteAdapter) CleanDBQueries() types.SQLCleanDBQuery {

	// Chain info
	selectChainIDQry := fmt.Sprintf(`
		SELECT 
		COUNT(*) REGISTERS,
		COALESCE(MAX(%s),'') CHAINID,
		COALESCE(MAX(%s),'') BVERSION 
		FROM %s;`,
		types.SQLColumnLabelChainID, types.SQLColumnLabelBurrowVer,
		types.SQLChainInfoTableName)

	deleteChainIDQry := fmt.Sprintf(`
		DELETE FROM %s;`,
		types.SQLChainInfoTableName)

	insertChainIDQry := fmt.Sprintf(`
		INSERT INTO %s (%s,%s) VALUES($1,$2)`,
		types.SQLChainInfoTableName,
		types.SQLColumnLabelChainID, types.SQLColumnLabelBurrowVer)

	// Dictionary
	selectDictionaryQry := fmt.Sprintf(`
		SELECT DISTINCT %s 
		FROM %s 
 		WHERE %s
		NOT IN ('%s','%s','%s');`,
		types.SQLColumnLabelTableName,
		types.SQLDictionaryTableName,
		types.SQLColumnLabelTableName,
		types.SQLLogTableName, types.SQLDictionaryTableName, types.SQLChainInfoTableName)

	deleteDictionaryQry := fmt.Sprintf(`
		DELETE FROM %s 
		WHERE %s 
		NOT IN ('%s','%s','%s');`,
		types.SQLDictionaryTableName,
		types.SQLColumnLabelTableName,
		types.SQLLogTableName, types.SQLDictionaryTableName, types.SQLChainInfoTableName)

	// log
	deleteLogQry := fmt.Sprintf(`
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
	return fmt.Sprintf(`DROP TABLE %s;`, tableName)
}
