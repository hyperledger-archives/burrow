package adapters

import (
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/vent/types"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/common/log"
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
	Schema string
	types.SQLNames
	Log *logging.Logger
}

var _ DBAdapter = &PostgresAdapter{}

// NewPostgresAdapter constructs a new db adapter
func NewPostgresAdapter(schema string, sqlNames types.SQLNames, log *logging.Logger) *PostgresAdapter {
	return &PostgresAdapter{
		Schema:   schema,
		SQLNames: sqlNames,
		Log:      log,
	}
}

func (pa *PostgresAdapter) Open(dbURL string) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", dbURL)
	if err != nil {
		log.Info("msg", "Error creating database connection", "err", err)
		return nil, err
	}

	if err := db.Ping(); err != nil {
		log.Info("msg", "Error opening database connection", "err", err)
		return nil, err
	}

	if pa.Schema != "" {
		err = ensureSchema(db, pa.Schema, pa.Log)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("no schema supplied")
	}

	return db, nil
}

func ensureSchema(db sqlx.Ext, schema string, log *logging.Logger) error {
	query := Cleanf(`SELECT EXISTS (SELECT 1 FROM pg_catalog.pg_namespace n WHERE n.nspname = '%s');`, schema)
	log.InfoMsg("FIND SCHEMA", "query", query)

	var found bool
	if err := db.QueryRowx(query).Scan(&found); err == nil {
		if !found {
			log.InfoMsg("Schema not found")
		}
		log.InfoMsg("Creating schema")

		query = Cleanf("CREATE SCHEMA %s;", schema)
		log.InfoMsg("CREATE SCHEMA", "query", query)

		if _, err = db.Exec(query); err != nil {
			if errorEquals(err, types.SQLErrorTypeDuplicatedSchema) {
				log.InfoMsg("Duplicated schema")
				return nil
			}
		}
	} else {
		log.InfoMsg("Error searching schema", "err", err)
		return err
	}
	return nil
}

// TypeMapping convert generic dataTypes to database dependent dataTypes
func (pa *PostgresAdapter) TypeMapping(sqlColumnType types.SQLColumnType) (string, error) {
	if sqlDataType, ok := pgDataTypes[sqlColumnType]; ok {
		return sqlDataType, nil
	}

	return "", fmt.Errorf("datatype %v not recognized", sqlColumnType)
}

// SecureColumnName return columns between appropriate security containers
func (pa *PostgresAdapter) SecureName(name string) string {
	return secureName(name)
}

// CreateTableQuery builds query for creating a new table
func (pa *PostgresAdapter) CreateTableQuery(tableName string, columns []*types.SQLTableColumn) (string, string) {
	// build query
	columnsDef := ""
	primaryKey := ""
	dictionaryValues := ""

	for i, column := range columns {
		secureColumn := pa.SecureName(column.Name)
		sqlType, _ := pa.TypeMapping(column.Type)
		pKey := 0

		if columnsDef != "" {
			columnsDef += ", "
			dictionaryValues += ", "
		}

		columnsDef += Cleanf("%s %s", secureColumn, sqlType)

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

	query := Cleanf("CREATE TABLE %s.%s (%s", pa.Schema, pa.SecureName(tableName), columnsDef)
	if primaryKey != "" {
		query += "," + Cleanf("CONSTRAINT %s_pkey PRIMARY KEY (%s)", tableName, primaryKey)
	}
	query += ");"

	dictionaryQuery := Cleanf("INSERT INTO %s.%s (%s,%s,%s,%s,%s,%s) VALUES %s;",
		pa.Schema, pa.Tables.Dictionary,
		pa.Columns.TableName, pa.Columns.ColumnName,
		pa.Columns.ColumnType, pa.Columns.ColumnLength,
		pa.Columns.PrimaryKey, pa.Columns.ColumnOrder,
		dictionaryValues)

	return query, dictionaryQuery
}

// FindTableQuery returns a query that checks if a table exists
func (pa *PostgresAdapter) FindTableQuery() string {
	query := "SELECT COUNT(*) found FROM %s.%s WHERE %s = $1;"

	return Cleanf(query,
		pa.Schema, pa.Tables.Dictionary, // from
		pa.Columns.TableName) // where
}

// TableDefinitionQuery returns a query with table structure
func (pa *PostgresAdapter) TableDefinitionQuery() string {
	query := `
		SELECT
			%s,%s,%s,%s
		FROM
			%s.%s
		WHERE
			%s = $1
		ORDER BY
			%s;`

	return Cleanf(query,
		pa.Columns.ColumnName, pa.Columns.ColumnType, // select
		pa.Columns.ColumnLength, pa.Columns.PrimaryKey, // select
		pa.Schema, pa.Tables.Dictionary, // from
		pa.Columns.TableName,   // where
		pa.Columns.ColumnOrder) // order by

}

// AlterColumnQuery returns a query for adding a new column to a table
func (pa *PostgresAdapter) AlterColumnQuery(tableName, columnName string, sqlColumnType types.SQLColumnType, length, order int) (string, string) {
	sqlType, _ := pa.TypeMapping(sqlColumnType)
	if length > 0 {
		sqlType = Cleanf("%s(%d)", sqlType, length)
	}

	query := Cleanf("ALTER TABLE %s.%s ADD COLUMN %s %s;",
		pa.Schema,
		pa.SecureName(tableName),
		pa.SecureName(columnName),
		sqlType)

	dictionaryQuery := Cleanf(`
		INSERT INTO %s.%s (%s,%s,%s,%s,%s,%s)
		VALUES ('%s','%s',%d,%d,%d,%d);`,

		pa.Schema, pa.Tables.Dictionary,

		pa.Columns.TableName, pa.Columns.ColumnName,
		pa.Columns.ColumnType, pa.Columns.ColumnLength,
		pa.Columns.PrimaryKey, pa.Columns.ColumnOrder,

		tableName, columnName, sqlColumnType, length, 0, order)

	return query, dictionaryQuery
}

// SelectRowQuery returns a query for selecting row values
func (pa *PostgresAdapter) SelectRowQuery(tableName, fields, indexValue string) string {
	return Cleanf("SELECT %s FROM %s.%s WHERE %s = '%s';",
		fields,                              // select
		pa.Schema, pa.SecureName(tableName), // from
		pa.Columns.Height, indexValue, // where
	)
}

// SelectLogQuery returns a query for selecting all tables involved in a block trn
func (pa *PostgresAdapter) SelectLogQuery() string {
	query := `
		SELECT DISTINCT %s,%s FROM %s.%s l WHERE %s = $1 AND %s = $2;`

	return Cleanf(query,
		pa.Columns.TableName, pa.Columns.EventName, // select
		pa.Schema, pa.Tables.Log, // from
		pa.Columns.Height,
		pa.Columns.ChainID) // where
}

// InsertLogQuery returns a query to insert a row in log table
func (pa *PostgresAdapter) InsertLogQuery() string {
	query := `
		INSERT INTO %s.%s (%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)
		VALUES (CURRENT_TIMESTAMP, $1, $2, $3, $4, $5, $6 ,$7, $8, $9, $10);`

	return Cleanf(query,
		pa.Schema, pa.Tables.Log, // insert
		//fields
		pa.Columns.TimeStamp,
		pa.Columns.ChainID, pa.Columns.TableName, pa.Columns.EventName, pa.Columns.EventFilter,
		pa.Columns.Height, pa.Columns.TxHash, pa.Columns.Action, pa.Columns.DataRow,
		pa.Columns.SqlStmt, pa.Columns.SqlValues)
}

// ErrorEquals verify if an error is of a given SQL type
func (pa *PostgresAdapter) ErrorEquals(err error, sqlErrorType types.SQLErrorType) bool {
	return errorEquals(err, sqlErrorType)
}

func errorEquals(err error, sqlErrorType types.SQLErrorType) bool {
	pqErr := new(pq.Error)

	if errors.As(err, &pqErr) {
		switch sqlErrorType {
		case types.SQLErrorTypeGeneric:
			return true
		case types.SQLErrorTypeDuplicatedColumn:
			return pqErr.Code == "42701"
		case types.SQLErrorTypeDuplicatedTable:
			return pqErr.Code == "42P07"
		case types.SQLErrorTypeDuplicatedSchema:
			return pqErr.Code == "42P06"
		case types.SQLErrorTypeUndefinedTable:
			return pqErr.Code == "42P01"
		case types.SQLErrorTypeUndefinedColumn:
			return pqErr.Code == "42703"
		case types.SQLErrorTypeInvalidType:
			return pqErr.Code == "42704"
		}
	}

	return false
}

func (pa *PostgresAdapter) UpsertQuery(table *types.SQLTable, row types.EventDataRow) (types.UpsertDeleteQuery, interface{}, error) {

	pointers := make([]interface{}, 0)

	columns := ""
	insValues := ""
	updValues := ""
	values := ""
	var txHash interface{} = nil

	i := 0

	// for each column in table
	for _, column := range table.Columns {
		secureColumn := pa.SecureName(column.Name)

		i++

		// INSERT INTO TABLE (*columns).........
		if columns != "" {
			columns += ", "
			insValues += ", "
			values += ", "
		}
		columns += secureColumn
		insValues += "$" + Cleanf("%d", i)

		// find data for column
		if value, ok := row.RowData[column.Name]; ok {
			// load hash value
			if column.Name == pa.Columns.TxHash {
				txHash = value
			}

			// column found (not null)
			// load values
			pointers = append(pointers, &value)
			values += fmt.Sprint(value)

			if !column.Primary {
				// column is not PK
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
			//pointers = append(pointers, &null)
			pointers = append(pointers, nil)
			values += "null"
		}
	}

	query := Cleanf("INSERT INTO %s.%s (%s) VALUES (%s) ", pa.Schema, pa.SecureName(table.Name),
		columns, insValues)

	if updValues != "" {
		query += Cleanf("ON CONFLICT ON CONSTRAINT %s_pkey DO UPDATE SET %s", table.Name, updValues)
	} else {
		query += Cleanf("ON CONFLICT ON CONSTRAINT %s_pkey DO NOTHING", table.Name)
	}
	query += ";"

	return types.UpsertDeleteQuery{Query: query, Values: values, Pointers: pointers}, txHash, nil
}

func (pa *PostgresAdapter) DeleteQuery(table *types.SQLTable, row types.EventDataRow) (types.UpsertDeleteQuery, error) {

	pointers := make([]interface{}, 0)
	columns := ""
	values := ""
	i := 0

	// for each column in table
	for _, column := range table.Columns {

		//only PK for delete
		if column.Primary {
			i++

			secureColumn := pa.SecureName(column.Name)

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

	query := Cleanf("DELETE FROM %s.%s WHERE %s;", pa.Schema, pa.SecureName(table.Name), columns)

	return types.UpsertDeleteQuery{Query: query, Values: values, Pointers: pointers}, nil
}

func (pa *PostgresAdapter) RestoreDBQuery() string {
	return Cleanf(`SELECT %s, %s, %s, %s, %s FROM %s 
								WHERE %s != '%s' AND  %s != '%s' AND to_char(%s,'YYYY-MM-DD HH24:MI:SS')<=$1 
								ORDER BY %s;`,
		pa.Columns.Id, pa.Columns.TableName, pa.Columns.Action, // select id, table, action
		pa.Columns.SqlStmt, pa.Columns.SqlValues, // select stmt, values
		pa.SchemaName(pa.Tables.Log),          // from
		pa.Columns.TableName, pa.Tables.Block, // where not _vent_block
		pa.Columns.TableName, pa.Tables.Tx, // where not _vent_tx
		pa.Columns.TimeStamp, // where time
		pa.Columns.Id)
}

func (pa *PostgresAdapter) CleanDBQueries() types.SQLCleanDBQuery {

	// Chain info
	selectChainIDQry := Cleanf(`
		SELECT 
		COUNT(*) REGISTERS,
		COALESCE(MAX(%s),'') CHAINID,
		COALESCE(MAX(%s),'') BVERSION 
		FROM %s.%s;`,
		pa.Columns.ChainID, pa.Columns.BurrowVersion,
		pa.Schema, pa.Tables.ChainInfo)

	deleteChainIDQry := Cleanf(`
		DELETE FROM %s;`,
		pa.SchemaName(pa.Tables.ChainInfo))

	insertChainIDQry := Cleanf(`
		INSERT INTO %s (%s,%s,%s) VALUES($1,$2,$3)`,
		pa.SchemaName(pa.Tables.ChainInfo),
		pa.Columns.ChainID, pa.Columns.BurrowVersion, pa.Columns.Height)

	// Dictionary
	selectDictionaryQry := Cleanf(`
		SELECT DISTINCT %s 
		FROM %s.%s 
 		WHERE %s
		NOT IN ('%s','%s','%s');`,
		pa.Columns.TableName,
		pa.Schema, pa.Tables.Dictionary,
		pa.Columns.TableName,
		pa.Tables.Log, pa.Tables.Dictionary, pa.Tables.ChainInfo)

	deleteDictionaryQry := Cleanf(`
		DELETE FROM %s.%s 
		WHERE %s 
		NOT IN ('%s','%s','%s');`,
		pa.Schema, pa.Tables.Dictionary,
		pa.Columns.TableName,
		pa.Tables.Log, pa.Tables.Dictionary, pa.Tables.ChainInfo)

	// log
	deleteLogQry := Cleanf(`
		DELETE FROM %s.%s;`,
		pa.Schema, pa.Tables.Log)

	return types.SQLCleanDBQuery{
		SelectChainIDQry:    selectChainIDQry,
		DeleteChainIDQry:    deleteChainIDQry,
		InsertChainIDQry:    insertChainIDQry,
		SelectDictionaryQry: selectDictionaryQry,
		DeleteDictionaryQry: deleteDictionaryQry,
		DeleteLogQry:        deleteLogQry,
	}
}

func (pa *PostgresAdapter) DropTableQuery(tableName string) string {
	// We cascade here to drop any associated views and triggers. We work under the assumption that vent
	// owns its database and any users need to be able to recreate objects that depend on vent tables in the event of
	// table drops
	return Cleanf(`DROP TABLE IF EXISTS %s CASCADE;`, pa.SchemaName(tableName))
}

func (pa *PostgresAdapter) CreateNotifyFunctionQuery(function, channel string, columns ...string) string {
	return Cleanf(`CREATE OR REPLACE FUNCTION %s() RETURNS trigger AS
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
		pa.SchemaName(function),                                         // create function
		channel, pa.Columns.Action, jsonBuildObjectArgs("OLD", columns), // case delete
		channel, pa.Columns.Action, jsonBuildObjectArgs("NEW", columns), // case else
	)
}

func (pa *PostgresAdapter) CreateTriggerQuery(triggerName, tableName, functionName string) string {
	trigger := pa.SecureName(triggerName)
	table := pa.SchemaName(tableName)
	return Cleanf(`DROP TRIGGER IF EXISTS %s ON %s CASCADE; 
		CREATE TRIGGER %s AFTER INSERT OR UPDATE OR DELETE ON %s
		FOR EACH ROW 
		EXECUTE PROCEDURE %s();
		`,
		trigger,                     // drop
		table,                       // on
		trigger,                     // create
		table,                       // on
		pa.SchemaName(functionName), // function
	)
}

func (pa *PostgresAdapter) SchemaName(tableName string) string {
	return fmt.Sprintf("%s.%s", pa.Schema, pa.SecureName(tableName))
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
