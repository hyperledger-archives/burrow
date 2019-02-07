package adapters

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/hyperledger/burrow/vent/logger"
	"github.com/hyperledger/burrow/vent/types"
	"github.com/lib/pq"
)

var pgDataTypes = map[types.SQLColumnType]string{
	types.SQLColumnTypeBool:      "BOOLEAN",
	types.SQLColumnTypeByteA:     "BYTEA",
	types.SQLColumnTypeInt:       "INTEGER",
	types.SQLColumnTypeSerial:    "SERIAL",
	types.SQLColumnTypeText:      "TEXT",
	types.SQLColumnTypeVarchar:   "VARCHAR",
	types.SQLColumnTypeTimeStamp: "TIMESTAMP",
	types.SQLColumnTypeNumeric:   "NUMERIC",
	types.SQLColumnTypeJSON:      "JSON",
	types.SQLColumnTypeBigInt:    "BIGINT",
}

// PostgresAdapter implements DBAdapter for Postgres
type PostgresAdapter struct {
	Log    *logger.Logger
	Schema string
}

// NewPostgresAdapter constructs a new db adapter
func NewPostgresAdapter(schema string, log *logger.Logger) *PostgresAdapter {
	return &PostgresAdapter{
		Log:    log,
		Schema: schema,
	}
}

// Open connects to a PostgreSQL database, opens it & create default schema if provided
func (adapter *PostgresAdapter) Open(dbURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		adapter.Log.Info("msg", "Error creating database connection", "err", err)
		return nil, err
	}

	// if there is a supplied Schema
	if adapter.Schema != "" {
		if err = db.Ping(); err != nil {
			adapter.Log.Info("msg", "Error opening database connection", "err", err)
			return nil, err
		}

		var found bool

		query := fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM pg_catalog.pg_namespace n WHERE n.nspname = '%s');`, adapter.Schema)
		adapter.Log.Info("msg", "FIND SCHEMA", "query", query)

		if err := db.QueryRow(query).Scan(&found); err == nil {
			if !found {
				adapter.Log.Warn("msg", "Schema not found")
			}
			adapter.Log.Info("msg", "Creating schema")

			query = fmt.Sprintf("CREATE SCHEMA %s;", adapter.Schema)
			adapter.Log.Info("msg", "CREATE SCHEMA", "query", query)

			if _, err = db.Exec(query); err != nil {
				if adapter.ErrorEquals(err, types.SQLErrorTypeDuplicatedSchema) {
					adapter.Log.Warn("msg", "Duplicated schema")
					return db, nil
				}
			}
		} else {
			adapter.Log.Info("msg", "Error searching schema", "err", err)
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("no schema supplied")
	}

	return db, err
}

// TypeMapping convert generic dataTypes to database dependent dataTypes
func (adapter *PostgresAdapter) TypeMapping(sqlColumnType types.SQLColumnType) (string, error) {
	if sqlDataType, ok := pgDataTypes[sqlColumnType]; ok {
		return sqlDataType, nil
	}

	return "", fmt.Errorf("datatype %v not recognized", sqlColumnType)
}

// SecureColumnName return columns between appropriate security containers
func (adapter *PostgresAdapter) SecureName(name string) string {
	return secureName(name)
}

// CreateTableQuery builds query for creating a new table
func (adapter *PostgresAdapter) CreateTableQuery(tableName string, columns []types.SQLTableColumn) (string, string) {
	// build query
	columnsDef := ""
	primaryKey := ""
	dictionaryValues := ""

	for i, tableColumn := range columns {
		secureColumn := adapter.SecureName(tableColumn.Name)
		sqlType, _ := adapter.TypeMapping(tableColumn.Type)
		pKey := 0

		if columnsDef != "" {
			columnsDef += ", "
			dictionaryValues += ", "
		}

		columnsDef += fmt.Sprintf("%s %s", secureColumn, sqlType)

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

	secureTable := adapter.SecureName(tableName)

	query := fmt.Sprintf("CREATE TABLE %s.%s (%s", adapter.Schema, secureTable, columnsDef)
	if primaryKey != "" {
		query += "," + fmt.Sprintf("CONSTRAINT %s_pkey PRIMARY KEY (%s)", tableName, primaryKey)
	}
	query += ");"

	dictionaryQuery := fmt.Sprintf("INSERT INTO %s.%s (%s,%s,%s,%s,%s,%s) VALUES %s;",
		adapter.Schema, types.SQLDictionaryTableName,
		types.SQLColumnLabelTableName, types.SQLColumnLabelColumnName,
		types.SQLColumnLabelColumnType, types.SQLColumnLabelColumnLength,
		types.SQLColumnLabelPrimaryKey, types.SQLColumnLabelColumnOrder,
		dictionaryValues)

	return query, dictionaryQuery
}

// LastBlockIDQuery returns a query for last inserted blockId in log table
func (adapter *PostgresAdapter) LastBlockIDQuery() string {
	query := `
		WITH ll AS (
			SELECT MAX(%s) AS %s FROM %s.%s
		)
		SELECT COALESCE(%s, '0') AS %s
			FROM ll LEFT OUTER JOIN %s.%s log ON (ll.%s = log.%s);`

	return fmt.Sprintf(query,
		types.SQLColumnLabelId,                // max
		types.SQLColumnLabelId,                // as
		adapter.Schema, types.SQLLogTableName, // from
		types.SQLColumnLabelHeight,            // coalesce
		types.SQLColumnLabelHeight,            // as
		adapter.Schema, types.SQLLogTableName, // from
		types.SQLColumnLabelId, types.SQLColumnLabelId) // on

}

// FindTableQuery returns a query that checks if a table exists
func (adapter *PostgresAdapter) FindTableQuery() string {
	query := "SELECT COUNT(*) found FROM %s.%s WHERE %s = $1;"

	return fmt.Sprintf(query,
		adapter.Schema, types.SQLDictionaryTableName, // from
		types.SQLColumnLabelTableName) // where
}

// TableDefinitionQuery returns a query with table structure
func (adapter *PostgresAdapter) TableDefinitionQuery() string {
	query := `
		SELECT
			%s,%s,%s,%s
		FROM
			%s.%s
		WHERE
			%s = $1
		ORDER BY
			%s;`

	return fmt.Sprintf(query,
		types.SQLColumnLabelColumnName, types.SQLColumnLabelColumnType, // select
		types.SQLColumnLabelColumnLength, types.SQLColumnLabelPrimaryKey, // select
		adapter.Schema, types.SQLDictionaryTableName, // from
		types.SQLColumnLabelTableName,   // where
		types.SQLColumnLabelColumnOrder) // order by

}

// AlterColumnQuery returns a query for adding a new column to a table
func (adapter *PostgresAdapter) AlterColumnQuery(tableName, columnName string, sqlColumnType types.SQLColumnType, length, order int) (string, string) {
	sqlType, _ := adapter.TypeMapping(sqlColumnType)
	if length > 0 {
		sqlType = fmt.Sprintf("%s(%d)", sqlType, length)
	}

	secureTable := adapter.SecureName(tableName)
	query := fmt.Sprintf("ALTER TABLE %s.%s ADD COLUMN %s %s;",
		adapter.Schema,
		secureTable,
		adapter.SecureName(columnName),
		sqlType)

	dictionaryQuery := fmt.Sprintf(`
		INSERT INTO %s.%s (%s,%s,%s,%s,%s,%s)
		VALUES ('%s','%s',%d,%d,%d,%d);`,

		adapter.Schema, types.SQLDictionaryTableName,

		types.SQLColumnLabelTableName, types.SQLColumnLabelColumnName,
		types.SQLColumnLabelColumnType, types.SQLColumnLabelColumnLength,
		types.SQLColumnLabelPrimaryKey, types.SQLColumnLabelColumnOrder,

		secureTable, columnName, sqlColumnType, length, 0, order)

	return query, dictionaryQuery
}

// SelectRowQuery returns a query for selecting row values
func (adapter *PostgresAdapter) SelectRowQuery(tableName, fields, indexValue string) string {
	return fmt.Sprintf("SELECT %s FROM %s.%s WHERE %s = '%s';", fields, adapter.Schema, tableName, types.SQLColumnLabelHeight, indexValue)
}

// SelectLogQuery returns a query for selecting all tables involved in a block trn
func (adapter *PostgresAdapter) SelectLogQuery() string {
	query := `
		SELECT DISTINCT %s,%s FROM %s.%s l WHERE %s = $1;`

	return fmt.Sprintf(query,
		types.SQLColumnLabelTableName, types.SQLColumnLabelEventName, // select
		adapter.Schema, types.SQLLogTableName, // from
		types.SQLColumnLabelHeight) // where
}

// InsertLogQuery returns a query to insert a row in log table
func (adapter *PostgresAdapter) InsertLogQuery() string {
	query := `
		INSERT INTO %s.%s (%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)
		VALUES (CURRENT_TIMESTAMP, $1, $2, $3, $4, $5, $6 ,$7, $8, $9);`

	return fmt.Sprintf(query,
		adapter.Schema, types.SQLLogTableName, // insert
		//fields
		types.SQLColumnLabelTimeStamp, types.SQLColumnLabelTableName, types.SQLColumnLabelEventName, types.SQLColumnLabelEventFilter,
		types.SQLColumnLabelHeight, types.SQLColumnLabelTxHash, types.SQLColumnLabelAction, types.SQLColumnLabelDataRow,
		types.SQLColumnLabelSqlStmt, types.SQLColumnLabelSqlValues)
}

// ErrorEquals verify if an error is of a given SQL type
func (adapter *PostgresAdapter) ErrorEquals(err error, sqlErrorType types.SQLErrorType) bool {
	if err, ok := err.(*pq.Error); ok {
		switch sqlErrorType {
		case types.SQLErrorTypeGeneric:
			return true
		case types.SQLErrorTypeDuplicatedColumn:
			return err.Code == "42701"
		case types.SQLErrorTypeDuplicatedTable:
			return err.Code == "42P07"
		case types.SQLErrorTypeDuplicatedSchema:
			return err.Code == "42P06"
		case types.SQLErrorTypeUndefinedTable:
			return err.Code == "42P01"
		case types.SQLErrorTypeUndefinedColumn:
			return err.Code == "42703"
		case types.SQLErrorTypeInvalidType:
			return err.Code == "42704"
		}
	}

	return false
}

func (adapter *PostgresAdapter) UpsertQuery(table types.SQLTable, row types.EventDataRow) (types.UpsertDeleteQuery, interface{}, error) {

	pointers := make([]interface{}, 0)

	columns := ""
	insValues := ""
	updValues := ""
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
			//pointers = append(pointers, &null)
			pointers = append(pointers, nil)
			values += "null"
		}
	}

	query := fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES (%s) ", adapter.Schema, table.Name, columns, insValues)

	if updValues != "" {
		query += fmt.Sprintf("ON CONFLICT ON CONSTRAINT %s_pkey DO UPDATE SET %s", table.Name, updValues)
	} else {
		query += fmt.Sprintf("ON CONFLICT ON CONSTRAINT %s_pkey DO NOTHING", table.Name)
	}
	query += ";"

	return types.UpsertDeleteQuery{Query: query, Values: values, Pointers: pointers}, txHash, nil
}

func (adapter *PostgresAdapter) DeleteQuery(table types.SQLTable, row types.EventDataRow) (types.UpsertDeleteQuery, error) {

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

	query := fmt.Sprintf("DELETE FROM %s.%s WHERE %s;", adapter.Schema, table.Name, columns)

	return types.UpsertDeleteQuery{Query: query, Values: values, Pointers: pointers}, nil
}

func (adapter *PostgresAdapter) RestoreDBQuery() string {
	return fmt.Sprintf(`SELECT %s, %s, %s, %s FROM %s.%s 
								WHERE to_char(%s,'YYYY-MM-DD HH24:MI:SS')<=$1 
								ORDER BY %s;`,
		types.SQLColumnLabelTableName, types.SQLColumnLabelAction, types.SQLColumnLabelSqlStmt, types.SQLColumnLabelSqlValues,
		adapter.Schema, types.SQLLogTableName,
		types.SQLColumnLabelTimeStamp,
		types.SQLColumnLabelId)
}

func (adapter *PostgresAdapter) CleanDBQueries() types.SQLCleanDBQuery {

	// Chain info
	selectChainIDQry := fmt.Sprintf(`
		SELECT 
		COUNT(*) REGISTERS,
		COALESCE(MAX(%s),'') CHAINID,
		COALESCE(MAX(%s),'') BVERSION 
		FROM %s.%s;`,
		types.SQLColumnLabelChainID, types.SQLColumnLabelBurrowVer,
		adapter.Schema, types.SQLChainInfoTableName)

	deleteChainIDQry := fmt.Sprintf(`
		DELETE FROM %s.%s;`,
		adapter.Schema, types.SQLChainInfoTableName)

	insertChainIDQry := fmt.Sprintf(`
		INSERT INTO %s.%s (%s,%s) VALUES($1,$2)`,
		adapter.Schema, types.SQLChainInfoTableName,
		types.SQLColumnLabelChainID, types.SQLColumnLabelBurrowVer)

	// Dictionary
	selectDictionaryQry := fmt.Sprintf(`
		SELECT DISTINCT %s 
		FROM %s.%s 
 		WHERE %s
		NOT IN ('%s','%s','%s');`,
		types.SQLColumnLabelTableName,
		adapter.Schema, types.SQLDictionaryTableName,
		types.SQLColumnLabelTableName,
		types.SQLLogTableName, types.SQLDictionaryTableName, types.SQLChainInfoTableName)

	deleteDictionaryQry := fmt.Sprintf(`
		DELETE FROM %s.%s 
		WHERE %s 
		NOT IN ('%s','%s','%s');`,
		adapter.Schema, types.SQLDictionaryTableName,
		types.SQLColumnLabelTableName,
		types.SQLLogTableName, types.SQLDictionaryTableName, types.SQLChainInfoTableName)

	// log
	deleteLogQry := fmt.Sprintf(`
		DELETE FROM %s.%s;`,
		adapter.Schema, types.SQLLogTableName)

	return types.SQLCleanDBQuery{
		SelectChainIDQry:    selectChainIDQry,
		DeleteChainIDQry:    deleteChainIDQry,
		InsertChainIDQry:    insertChainIDQry,
		SelectDictionaryQry: selectDictionaryQry,
		DeleteDictionaryQry: deleteDictionaryQry,
		DeleteLogQry:        deleteLogQry,
	}
}

func (adapter *PostgresAdapter) DropTableQuery(tableName string) string {
	// We cascade here to drop any associated views and triggers. We work under the assumption that vent
	// owns its database and any users need to be able to recreate objects that depend on vent tables in the event of
	// table drops
	return fmt.Sprintf(`DROP TABLE %s.%s CASCADE;`, adapter.Schema, tableName)
}

func (adapter *PostgresAdapter) CreateNotifyFunctionQuery(function, channel string, columns ...string) string {
	return fmt.Sprintf(`CREATE OR REPLACE FUNCTION %s.%s() RETURNS trigger AS
		$trigger$
		BEGIN
			CASE TG_OP
			WHEN 'DELETE' THEN
				PERFORM pg_notify('%s', CAST(json_build_object('%s', TG_OP, %s) as text));
				RETURN OLD;
			ELSE
				PERFORM pg_notify('%s', CAST(json_build_object('%s', TG_OP, %s) as text));
				RETURN NEW;
			END CASE;
		END;
		$trigger$
		LANGUAGE 'plpgsql';
		`,
		adapter.Schema, function, // create function
		channel, types.SQLColumnLabelAction, jsonBuildObjectArgs("OLD", columns), // case delete
		channel, types.SQLColumnLabelAction, jsonBuildObjectArgs("NEW", columns), // case else
	)
}

func (adapter *PostgresAdapter) CreateTriggerQuery(triggerName, tableName, functionName string) string {
	return fmt.Sprintf(`CREATE TRIGGER %s AFTER INSERT OR UPDATE OR DELETE ON %s.%s
		FOR EACH ROW 
		EXECUTE FUNCTION %s.%s();
		`,
		triggerName,
		adapter.Schema, tableName,
		adapter.Schema, functionName,
	)
}

func secureName(columnName string) string {
	return `"` + columnName + `"`
}

func jsonBuildObjectArgs(record string, columns []string) string {
	elements := make([]string, len(columns))
	for i, column := range columns {
		elements[i] = "'" + column + "', " + record + "." + secureName(column)
	}

	return strings.Join(elements, ", ")
}
