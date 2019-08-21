package sqldb

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/vent/sqldb/adapters"
	"github.com/hyperledger/burrow/vent/types"
	"github.com/jmoiron/sqlx"
)

const maxUint64 uint64 = (1 << 64) - 1

var tables = types.DefaultSQLTableNames
var columns = types.DefaultSQLColumnNames

// SQLDB implements the access to a sql database
type SQLDB struct {
	DB *sqlx.DB
	adapters.DBAdapter
	Schema  string
	Queries Queries
	types.SQLNames
	Log *logging.Logger
}

// NewSQLDB delegates work to a specific database adapter implementation,
// opens database connection and create log tables
func NewSQLDB(connection types.SQLConnection) (*SQLDB, error) {
	db := &SQLDB{
		Schema:   connection.DBSchema,
		SQLNames: types.DefaultSQLNames,
		Log:      connection.Log,
	}

	switch connection.DBAdapter {
	case types.PostgresDB:
		db.DBAdapter = adapters.NewPostgresAdapter(safe(connection.DBSchema), db.SQLNames, connection.Log)

	case types.SQLiteDB:
		db.DBAdapter = adapters.NewSQLiteAdapter(db.SQLNames, connection.Log)
	default:
		return nil, errors.New("invalid database adapter")
	}

	var err error
	db.DB, err = db.DBAdapter.Open(connection.DBURL)
	if err != nil {
		db.Log.InfoMsg("Error opening database connection", "err", err)
		return nil, err
	}

	if err = db.Ping(); err != nil {
		db.Log.InfoMsg("Error database not available", "err", err)
		return nil, err
	}

	return db, nil
}

// Initialise the system and chain tables in case this is the first run - is idempotent though will drop tables
// if ChainID has changed
func (db *SQLDB) Init(chainID, burrowVersion string) error {
	db.Log.InfoMsg("Initializing DB")

	// Create dictionary and log tables
	sysTables := db.systemTablesDefinition()

	// IMPORTANT: DO NOT CHANGE TABLE CREATION ORDER (1)
	if err := db.createTable(chainID, sysTables[db.Tables.Dictionary], true); err != nil {
		if !db.DBAdapter.ErrorEquals(err, types.SQLErrorTypeDuplicatedTable) {
			db.Log.InfoMsg("Error creating Dictionary table", "err", err)
			return err
		}
	}

	// IMPORTANT: DO NOT CHANGE TABLE CREATION ORDER (2)
	if err := db.createTable(chainID, sysTables[db.Tables.Log], true); err != nil {
		if !db.DBAdapter.ErrorEquals(err, types.SQLErrorTypeDuplicatedTable) {
			db.Log.InfoMsg("Error creating Log table", "err", err)
			return err
		}
	}

	// IMPORTANT: DO NOT CHANGE TABLE CREATION ORDER (3)
	if err := db.createTable(chainID, sysTables[db.Tables.ChainInfo], true); err != nil {
		if !db.DBAdapter.ErrorEquals(err, types.SQLErrorTypeDuplicatedTable) {
			db.Log.InfoMsg("Error creating Chain Info table", "err", err)
			return err
		}
	}

	chainIDChanged, err := db.InitChain(chainID, burrowVersion)
	if err != nil {
		return fmt.Errorf("could not initialise chain in database: %v", err)
	}

	if chainIDChanged {
		// If the chain has changed - drop existing data
		err = db.CleanTables(chainID, burrowVersion)
		if err != nil {
			return fmt.Errorf("could not clean tables after ChainID change: %v", err)
		}
	}

	db.Queries, err = db.prepareQueries()
	if err != nil {
		db.Log.InfoMsg("Could not prepare queries", "err", err)
		return err
	}

	return nil
}

func (db *SQLDB) prepareQueries() (Queries, error) {
	err := new(error)
	//language=SQL
	return Queries{
		LastBlockHeight: db.prepare(err, fmt.Sprintf("SELECT %s FROM %s WHERE %s=:chainid",
			db.Columns.Height, // select
			db.DBAdapter.SchemaName(db.Tables.ChainInfo), // from
			db.DBAdapter.SecureName(db.Columns.ChainID),  // where
		)),
		SetBlockHeight: fmt.Sprintf("UPDATE %s SET %s = :height WHERE %s = :chainid",
			db.DBAdapter.SchemaName(db.Tables.ChainInfo), // update
			db.Columns.Height,  // set
			db.Columns.ChainID, // where
		),
	}, *err
}

func (db *SQLDB) InitChain(chainID, burrowVersion string) (chainIDChanged bool, _ error) {
	cleanQueries := db.DBAdapter.CleanDBQueries()

	var savedChainID, savedBurrowVersion, query string
	savedRows := 0

	// Read chainID
	query = cleanQueries.SelectChainIDQry
	if err := db.DB.QueryRow(query).Scan(&savedRows, &savedChainID, &savedBurrowVersion); err != nil {
		db.Log.InfoMsg("Error selecting CHAIN ID", "err", err, "query", query)
		return false, err
	}

	if savedRows == 1 {
		return savedChainID != chainID, nil
	}

	if savedRows > 1 {
		return false, fmt.Errorf("error multiple chains defined returned")
	}

	// First database access
	// Save new values and exit
	query = cleanQueries.InsertChainIDQry
	_, err := db.DB.Exec(query, chainID, burrowVersion, 0)

	if err != nil {
		db.Log.InfoMsg("Error inserting CHAIN ID", "err", err, "query", query)
	}
	return false, err

}

// CleanTables drop tables if stored chainID is different from the given one & store new chainID
// if the chainID is the same, do nothing
func (db *SQLDB) CleanTables(chainID, burrowVersion string) error {
	var tx *sql.Tx
	var err error
	var tableName string
	tables := make([]string, 0)
	cleanQueries := db.DBAdapter.CleanDBQueries()

	// Begin tx
	if tx, err = db.DB.Begin(); err != nil {
		db.Log.InfoMsg("Error beginning transaction", "err", err)
		return err
	}
	defer tx.Rollback()

	// Delete chainID
	query := cleanQueries.DeleteChainIDQry
	if _, err = tx.Exec(query); err != nil {
		db.Log.InfoMsg("Error deleting CHAIN ID", "err", err, "query", query)
		return err
	}

	// Insert chainID
	query = cleanQueries.InsertChainIDQry
	if _, err := tx.Exec(query, chainID, burrowVersion, 0); err != nil {
		db.Log.InfoMsg("Error inserting CHAIN ID", "err", err, "query", query)
		return err
	}

	// Load Tables
	query = cleanQueries.SelectDictionaryQry
	rows, err := tx.Query(query)
	if err != nil {
		db.Log.InfoMsg("error querying dictionary", "err", err, "query", query)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(&tableName); err != nil {
			db.Log.InfoMsg("error scanning table structure", "err", err)
			return err
		}

		if err = rows.Err(); err != nil {
			db.Log.InfoMsg("error scanning table structure", "err", err)
			return err
		}
		tables = append(tables, tableName)
	}

	// Delete Dictionary
	query = cleanQueries.DeleteDictionaryQry
	if _, err = tx.Exec(query); err != nil {
		db.Log.InfoMsg("Error deleting dictionary", "err", err, "query", query)
		return err
	}

	// Delete Log
	query = cleanQueries.DeleteLogQry
	if _, err = tx.Exec(query); err != nil {
		db.Log.InfoMsg("Error deleting log", "err", err, "query", query)
		return err
	}
	// Drop database tables
	for _, tableName = range tables {
		query = db.DBAdapter.DropTableQuery(tableName)
		if _, err = tx.Exec(query); err != nil {
			// if error == table does not exists, continue
			if !db.DBAdapter.ErrorEquals(err, types.SQLErrorTypeUndefinedTable) {
				db.Log.InfoMsg("error dropping tables", "err", err, "value", tableName, "query", query)
				return err
			}
		}
	}

	// Commit
	if err = tx.Commit(); err != nil {
		db.Log.InfoMsg("Error commiting transaction", "err", err)
		return err
	}

	return nil
}

// Close database connection
func (db *SQLDB) Close() {
	if err := db.DB.Close(); err != nil {
		db.Log.InfoMsg("Error closing database", "err", err)
	}
}

// Ping database
func (db *SQLDB) Ping() error {
	if err := db.DB.Ping(); err != nil {
		db.Log.InfoMsg("Error database not available", "err", err)
		return err
	}

	return nil
}

// SynchronizeDB synchronize db tables structures from given tables specifications
func (db *SQLDB) SynchronizeDB(chainID string, eventTables types.EventTables) error {
	db.Log.InfoMsg("Synchronizing DB")

	for _, table := range eventTables {
		found, err := db.findTable(table.Name)
		if err != nil {
			return err
		}

		if found {
			err = db.alterTable(chainID, table)
		} else {
			err = db.createTable(chainID, table, false)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// SetBlock inserts or updates multiple rows and stores log info in SQL tables
func (db *SQLDB) SetBlock(chainID string, eventTables types.EventTables, eventData types.EventData) error {
	db.Log.InfoMsg("Synchronize Block", "action", "SYNC")

	// Begin tx
	tx, err := db.DB.Beginx()
	if err != nil {
		db.Log.InfoMsg("Error beginning transaction", "err", err)
		return err
	}
	defer tx.Rollback()

	// Prepare log statement
	logQuery := db.DBAdapter.InsertLogQuery()
	logStmt, err := tx.Prepare(logQuery)
	if err != nil {
		db.Log.InfoMsg("Error preparing log stmt", "err", err)
		return err
	}
	defer logStmt.Close()

	var tableName string
loop:
	// for each table in the block
	for _, table := range eventTables {
		tableName = safe(table.Name)
		dataRows := eventData.Tables[table.Name]
		// for Each Row
		for _, row := range dataRows {
			var queryVal types.UpsertDeleteQuery
			var txHash interface{}
			var errQuery error

			switch row.Action {
			case types.ActionUpsert:
				//Prepare Upsert
				if queryVal, txHash, errQuery = db.DBAdapter.UpsertQuery(table, row); errQuery != nil {
					db.Log.InfoMsg("Error building upsert query", "err", errQuery, "value", fmt.Sprintf("%v %v", table, row))
					break loop // exits from all loops -> continue in close log stmt
				}

			case types.ActionDelete:
				//Prepare Delete
				if queryVal, errQuery = db.DBAdapter.DeleteQuery(table, row); errQuery != nil {
					db.Log.InfoMsg("Error building delete query", "err", errQuery, "value", fmt.Sprintf("%v %v", table, row))
					break loop // exits from all loops -> continue in close log stmt
				}
			default:
				//Invalid Action
				db.Log.InfoMsg("invalid action", "value", row.Action)
				err = fmt.Errorf("invalid row action %s", row.Action)
				break loop // exits from all loops -> continue in close log stmt
			}

			sqlQuery := queryVal.Query

			// Perform row action
			db.Log.InfoMsg("msg", "action", row.Action, "query", sqlQuery, "value", queryVal.Values)
			if _, err = tx.Exec(sqlQuery, queryVal.Pointers...); err != nil {
				db.Log.InfoMsg(fmt.Sprintf("error performing %s on row", row.Action), "err", err, "value", queryVal.Values)
				break loop // exits from all loops -> continue in close log stmt
			}

			// Marshal the rowData map
			rowData, err := getJSON(row.RowData)
			if err != nil {
				db.Log.InfoMsg("error marshaling rowData", "err", err, "value", fmt.Sprintf("%v", row.RowData))
				break loop // exits from all loops -> continue in close log stmt
			}

			// Marshal sql values
			sqlValues, err := getJSONFromValues(queryVal.Pointers)
			if err != nil {
				db.Log.InfoMsg("error marshaling rowdata", "err", err, "value", fmt.Sprintf("%v", row.RowData))
				break loop // exits from all loops -> continue in close log stmt
			}

			eventName, _ := row.RowData[db.Columns.EventName].(string)
			// Insert in log
			db.Log.InfoMsg("INSERT LOG",
				"log_query", logQuery,
				"chain_id", chainID,
				"table_name", tableName,
				"event_name", eventName,
				"event_filter", row.EventClass.GetFilter(),
				"block_height", eventData.BlockHeight,
				"tx_hash", txHash,
				"row_action", row.Action,
				"row_data", rowData,
				"sql_query", sqlQuery,
				"sql_values", sqlValues,
			)

			if _, err = logStmt.Exec(chainID, tableName, eventName, row.EventClass.GetFilter(), eventData.BlockHeight, txHash,
				row.Action, rowData, sqlQuery, sqlValues); err != nil {
				db.Log.InfoMsg("Error inserting into log", "err", err)
				break loop // exits from all loops -> continue in close log stmt
			}
		}
	}

	// Close log statement
	if err == nil {
		if err = logStmt.Close(); err != nil {
			db.Log.InfoMsg("Error closing log stmt", "err", err)
		}
	}

	// Error handling
	if err != nil {
		// Rollback error
		if errRb := tx.Rollback(); errRb != nil {
			db.Log.InfoMsg("Error on rollback", "err", errRb)
			return errRb
		}

		//Is a SQL error
		if db.DBAdapter.ErrorEquals(err, types.SQLErrorTypeGeneric) {

			// Table does not exists
			if db.DBAdapter.ErrorEquals(err, types.SQLErrorTypeUndefinedTable) {
				db.Log.InfoMsg("Table not found", "value", tableName)
				//Synchronize DB
				if err = db.SynchronizeDB(chainID, eventTables); err != nil {
					return err
				}
				//Retry
				return db.SetBlock(chainID, eventTables, eventData)
			}

			// Columns do not match
			if db.DBAdapter.ErrorEquals(err, types.SQLErrorTypeUndefinedColumn) {
				db.Log.InfoMsg("Column not found", "value", tableName)
				//Synchronize DB
				if err = db.SynchronizeDB(chainID, eventTables); err != nil {
					return err
				}
				//Retry
				return db.SetBlock(chainID, eventTables, eventData)
			}
			return err
		}
		return err
	}

	db.Log.InfoMsg("COMMIT", "action", "COMMIT")

	err = db.SetBlockHeight(tx, chainID, eventData.BlockHeight)
	if err != nil {
		db.Log.InfoMsg("Could not commit block height", "err", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		db.Log.InfoMsg("Error on commit", "err", err)
		return err
	}

	return nil
}

// GetBlock returns all tables structures and row data for given block
func (db *SQLDB) GetBlock(chainID string, height uint64) (types.EventData, error) {
	var data types.EventData
	data.BlockHeight = height
	data.Tables = make(map[string]types.EventDataTable)

	// get all table structures involved in the block
	tables, err := db.getBlockTables(chainID, height)
	if err != nil {
		return data, err
	}

	query := ""

	// for each table
	for _, table := range tables {
		// get query for table
		query, err = db.getSelectQuery(table, height)
		if err != nil {
			db.Log.InfoMsg("Error building table query", "err", err)
			return data, err
		}
		db.Log.InfoMsg("Query table data", "query", query)
		rows, err := db.DB.Query(query)
		if err != nil {
			db.Log.InfoMsg("Error querying table data", "err", err)
			return data, err
		}
		defer rows.Close()

		cols, err := rows.Columns()
		if err != nil {
			db.Log.InfoMsg("Error getting row columns", "err", err)
			return data, err
		}

		// builds pointers
		length := len(cols)
		pointers := make([]interface{}, length)
		containers := make([]sql.NullString, length)

		for i := range pointers {
			pointers[i] = &containers[i]
		}

		// for each row in table
		var dataRows []types.EventDataRow

		for rows.Next() {

			row := make(map[string]interface{})

			if err = rows.Scan(pointers...); err != nil {
				db.Log.InfoMsg("Error scanning data", "err", err)
				return data, err
			}
			db.Log.InfoMsg("Query resultset", "value", fmt.Sprintf("%+v", containers))

			// for each column in row
			for i, col := range cols {
				// add value if not null
				if containers[i].Valid {
					row[col] = containers[i].String
				}
			}
			dataRows = append(dataRows, types.EventDataRow{Action: types.ActionRead, RowData: row})
		}

		if err = rows.Err(); err != nil {
			db.Log.InfoMsg("Error during rows iteration", "err", err)
			return data, err
		}
		data.Tables[table.Name] = dataRows
	}
	return data, nil
}

func (db *SQLDB) LastBlockHeight(chainID string) (uint64, error) {
	const errHeader = "LastBlockHeight()"
	type arg struct {
		ChainID string
	}
	height := new(uint64)
	err := db.Queries.LastBlockHeight.Get(height, arg{ChainID: chainID})
	if err != nil {
		return 0, fmt.Errorf("%s: %v", errHeader, err)
	}
	return *height, nil
}

func (db *SQLDB) SetBlockHeight(tx sqlx.Ext, chainID string, height uint64) error {
	const errHeader = "SetBlockHeight()"
	type arg struct {
		ChainID string
		Height  uint64
	}
	result, err := sqlx.NamedExec(tx, db.Queries.SetBlockHeight, arg{ChainID: chainID, Height: height})
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: could not get rows affected: %v", errHeader, err)
	}
	if rows != 1 {
		return fmt.Errorf("%s: should update the height in  exactly one row for ChainID '%s' but %d rows were affected",
			errHeader, chainID, rows)
	}
	return nil
}

// RestoreDB restores the DB to a given moment in time. If prefix is provided restores the table state to a new set of
// tables as <prefix>_<table name>. Drops destination tables before recreating them. If zero time passed restores
// all values
func (db *SQLDB) RestoreDB(restoreTime time.Time, prefix string) error {
	const year = time.Hour * 24 * 365
	const yymmddhhmmss = "2006-01-02 15:04:05"

	var pointers []interface{}

	// Get Restore DB query
	query := db.DBAdapter.RestoreDBQuery()
	if restoreTime.IsZero() {
		// We'll assume a sufficiently small clock skew...
		restoreTime = time.Now().Add(100 * year)
	}
	strTime := restoreTime.Format(yymmddhhmmss)

	db.Log.InfoMsg("RESTORING DB..................................")

	db.Log.InfoMsg("open log", "query", query)
	// Postgres does not work is run within same tx as updates, see: https://github.com/lib/pq/issues/81
	rows, err := db.DB.Query(query, strTime)
	if err != nil {
		db.Log.InfoMsg("error querying log", "err", err)
		return err
	}
	defer rows.Close()

	tx, err := db.DB.Begin()
	if err != nil {
		db.Log.InfoMsg("could not open transaction for restore", "err", err)
		return err
	}
	defer tx.Rollback()

	// For each row returned
	for rows.Next() {
		var id int64
		var tableName, sqlSmt, sqlValues string
		var action types.DBAction

		err = rows.Err()
		if err != nil {
			db.Log.InfoMsg("error scanning table structure", "err", err)
			return err
		}

		err = rows.Scan(&id, &tableName, &action, &sqlSmt, &sqlValues)
		if err != nil {
			db.Log.InfoMsg("error scanning table structure", "err", err)
			return err
		}

		restoreTable := tableName
		if prefix != "" {
			restoreTable = fmt.Sprintf("%s_%s", prefix, tableName)
		}

		switch action {
		case types.ActionUpsert, types.ActionDelete:
			// get row values
			if pointers, err = getValuesFromJSON(sqlValues); err != nil {
				db.Log.InfoMsg("error unmarshaling json", "err", err, "value", sqlValues)
				return err
			}

			// Prepare Upsert/delete
			query = sqlSmt
			if prefix != "" {
				// TODO: [Silas] ugh this is a little fragile
				query = strings.Replace(sqlSmt, tableName, restoreTable, -1)
			}

			db.Log.InfoMsg("SQL COMMAND", "sql", query, "log_id", id)
			if _, err = tx.Exec(query, pointers...); err != nil {
				db.Log.InfoMsg("Error executing upsert/delete ", "err", err, "value", sqlSmt, "data", sqlValues)
				return err
			}

		case types.ActionAlterTable, types.ActionCreateTable:
			if action == types.ActionCreateTable {
				dropQuery := db.DBAdapter.DropTableQuery(restoreTable)
				_, err := tx.Exec(dropQuery)
				if err != nil && !db.DBAdapter.ErrorEquals(err, types.SQLErrorTypeUndefinedTable) {
					return fmt.Errorf("could not drop target restore table %s: %v", restoreTable, err)
				}
			}
			// Prepare Alter/Create Table
			query = strings.Replace(sqlSmt, tableName, restoreTable, -1)

			db.Log.InfoMsg("SQL COMMAND", "sql", query)
			_, err = tx.Exec(query)
			if err != nil {
				db.Log.InfoMsg("Error executing alter/create table command ", "err", err, "value", sqlSmt)
				return err
			}
		default:
			// Invalid Action
			db.Log.InfoMsg("invalid action", "value", action)
			return fmt.Errorf("invalid row action %s", action)
		}
	}
	err = tx.Commit()
	if err != nil {
		db.Log.InfoMsg("could not commit restore tx", "err", err)
		return err
	}
	return nil
}

func (db *SQLDB) prepare(perr *error, query string) *sqlx.NamedStmt {
	stmt, err := db.DB.PrepareNamed(query)
	if err != nil && *perr == nil {
		*perr = fmt.Errorf("could not prepare query '%s': %v", query, err)
	}
	return stmt
}

// findTable checks if a table exists in the default schema
func (db *SQLDB) findTable(tableName string) (bool, error) {

	found := 0
	safeTable := safe(tableName)
	query := db.DBAdapter.FindTableQuery()

	db.Log.InfoMsg("FIND TABLE", "query", query, "value", safeTable)
	if err := db.DB.QueryRow(query, tableName).Scan(&found); err != nil {
		db.Log.InfoMsg("Error finding table", "err", err)
		return false, err
	}

	if found == 0 {
		db.Log.InfoMsg("Table not found", "value", safeTable)
		return false, nil
	}

	return true, nil
}

// getTableDef returns the structure of a given SQL table
func (db *SQLDB) getTableDef(tableName string) (*types.SQLTable, error) {
	table := &types.SQLTable{
		Name: safe(tableName),
	}
	found, err := db.findTable(table.Name)
	if err != nil {
		return nil, err
	}

	if !found {
		db.Log.InfoMsg("Error table not found", "value", table.Name)
		return nil, errors.New("Error table not found " + table.Name)
	}

	query := db.DBAdapter.TableDefinitionQuery()

	db.Log.InfoMsg("QUERY STRUCTURE", "query", query, "value", table.Name)
	rows, err := db.DB.Query(query, safe(tableName))
	if err != nil {
		db.Log.InfoMsg("Error querying table structure", "err", err)
		return nil, err
	}
	defer rows.Close()

	var columns []*types.SQLTableColumn

	for rows.Next() {
		var columnName string
		var columnSQLType types.SQLColumnType
		var columnIsPK int
		var columnLength int

		if err = rows.Scan(&columnName, &columnSQLType, &columnLength, &columnIsPK); err != nil {
			db.Log.InfoMsg("Error scanning table structure", "err", err)
			return nil, err
		}

		if _, err = db.DBAdapter.TypeMapping(columnSQLType); err != nil {
			return nil, err
		}

		columns = append(columns, &types.SQLTableColumn{
			Name:    columnName,
			Type:    columnSQLType,
			Length:  columnLength,
			Primary: columnIsPK == 1,
		})
	}

	if err = rows.Err(); err != nil {
		db.Log.InfoMsg("Error during rows iteration", "err", err)
		return nil, err
	}

	table.Columns = columns
	return table, nil
}

// alterTable alters the structure of a SQL table & add info to the dictionary
func (db *SQLDB) alterTable(chainID string, table *types.SQLTable) error {
	db.Log.InfoMsg("Altering table", "value", table.Name)

	// prepare log query
	logQuery := db.DBAdapter.InsertLogQuery()

	// current table structure
	safeTable := safe(table.Name)
	currentTable, err := db.getTableDef(safeTable)
	if err != nil {
		return err
	}

	sqlValues, _ := getJSON(nil)

	// for each column in the new table structure
	for order, newColumn := range table.Columns {
		found := false

		// check if exists in the current table structure
		for _, currentColumn := range currentTable.Columns {
			// if column exists
			if currentColumn.Name == newColumn.Name {
				found = true
				break
			}
		}

		if !found {
			safeCol := safe(newColumn.Name)
			query, dictionary := db.DBAdapter.AlterColumnQuery(safeTable, safeCol, newColumn.Type, newColumn.Length, order)

			//alter column
			db.Log.InfoMsg("ALTER TABLE", "query", safe(query))
			_, err = db.DB.Exec(safe(query))

			if err != nil {
				if db.DBAdapter.ErrorEquals(err, types.SQLErrorTypeDuplicatedColumn) {
					db.Log.InfoMsg("Duplicate column", "value", safeCol)
				} else {
					db.Log.InfoMsg("Error altering table", "err", err)
					return err
				}
			} else {
				//store dictionary
				db.Log.InfoMsg("STORE DICTIONARY", "query", dictionary)
				_, err = db.DB.Exec(dictionary)
				if err != nil {
					db.Log.InfoMsg("Error storing  dictionary", "err", err)
					return err
				}

				// Marshal the table into a JSON string.
				var jsonData []byte
				jsonData, err = getJSON(newColumn)
				if err != nil {
					db.Log.InfoMsg("error marshaling column", "err", err, "value", fmt.Sprintf("%v", newColumn))
					return err
				}
				//insert log
				_, err = db.DB.Exec(logQuery, chainID, table.Name, "", "", nil, nil, types.ActionAlterTable, jsonData, query, sqlValues)
				if err != nil {
					db.Log.InfoMsg("Error inserting log", "err", err)
					return err
				}
			}
		}
	}

	// Ensure triggers are defined
	err = db.createTableTriggers(table)
	if err != nil {
		db.Log.InfoMsg("error creating notification triggers", "err", err, "value", fmt.Sprintf("%v", table))
		return fmt.Errorf("could not create table notification triggers: %v", err)
	}
	return nil
}

// createTable creates a new table
func (db *SQLDB) createTable(chainID string, table *types.SQLTable, isInitialise bool) error {
	db.Log.InfoMsg("Creating Table", "value", table.Name)

	// prepare log query
	logQuery := db.DBAdapter.InsertLogQuery()

	//get create table query
	safeTable := safe(table.Name)
	query, dictionary := db.DBAdapter.CreateTableQuery(safeTable, table.Columns)
	if query == "" {
		db.Log.InfoMsg("empty CREATE TABLE query")
		return errors.New("empty CREATE TABLE query")
	}

	// create table
	db.Log.InfoMsg("CREATE TABLE", "query", query)
	_, err := db.DB.Exec(query)
	if err != nil {
		return err
	}

	//store dictionary
	db.Log.InfoMsg("STORE DICTIONARY", "query", dictionary)
	_, err = db.DB.Exec(dictionary)
	if err != nil {
		db.Log.InfoMsg("Error storing  dictionary", "err", err)
		return err
	}

	err = db.createTableTriggers(table)
	if err != nil {
		db.Log.InfoMsg("error creating notification triggers", "err", err, "value", fmt.Sprintf("%v", table))
		return fmt.Errorf("could not create table notification triggers: %v", err)
	}

	//insert log (if action is not database initialization)
	if !isInitialise {
		// Marshal the table into a JSON string.
		var jsonData []byte
		jsonData, err = getJSON(table)
		if err != nil {
			db.Log.InfoMsg("error marshaling table", "err", err, "value", fmt.Sprintf("%v", table))
			return err
		}
		sqlValues, _ := getJSON(nil)

		//insert log
		_, err = db.DB.Exec(logQuery, chainID, table.Name, "", "", nil, nil, types.ActionCreateTable, jsonData, query, sqlValues)
		if err != nil {
			db.Log.InfoMsg("Error inserting log", "err", err)
			return err
		}
	}
	return nil
}

// Creates (or updates) table notification triggers and functions
func (db *SQLDB) createTableTriggers(table *types.SQLTable) error {
	// If the adapter supports notification triggers
	dbNotify, ok := db.DBAdapter.(adapters.DBNotifyTriggerAdapter)
	if ok {
		for channel, columns := range table.NotifyChannels {
			function := fmt.Sprintf("%s_%s_notify_function", table.Name, channel)

			query := dbNotify.CreateNotifyFunctionQuery(function, channel, columns...)
			db.Log.InfoMsg("CREATE NOTIFICATION FUNCTION", "query", query)
			_, err := db.DB.Exec(query)
			if err != nil {
				return fmt.Errorf("could not create notification function: %v", err)
			}

			trigger := fmt.Sprintf("%s_%s_notify_trigger", table.Name, channel)
			query = dbNotify.CreateTriggerQuery(trigger, table.Name, function)
			db.Log.InfoMsg("CREATE NOTIFICATION TRIGGER", "query", query)
			_, err = db.DB.Exec(query)
			if err != nil {
				return fmt.Errorf("could not create notification trigger: %v", err)
			}
		}
	}
	return nil
}

// getSelectQuery builds a select query for a specific SQL table and a given block
func (db *SQLDB) getSelectQuery(table *types.SQLTable, height uint64) (string, error) {

	fields := ""

	for _, tableColumn := range table.Columns {
		if fields != "" {
			fields += ", "
		}
		fields += db.DBAdapter.SecureName(tableColumn.Name)
	}

	if fields == "" {
		return "", errors.New("error table does not contain any fields")
	}

	query := db.DBAdapter.SelectRowQuery(table.Name, fields, strconv.FormatUint(height, 10))
	return query, nil
}

// getBlockTables return all SQL tables that have been involved
// in a given batch transaction for a specific block
func (db *SQLDB) getBlockTables(chainid string, height uint64) (types.EventTables, error) {
	tables := make(types.EventTables)

	query := db.DBAdapter.SelectLogQuery()
	db.Log.InfoMsg("QUERY LOG", "query", query, "height", height, "chainid", chainid)

	rows, err := db.DB.Query(query, height, chainid)
	if err != nil {
		db.Log.InfoMsg("Error querying log", "err", err)
		return tables, err
	}

	defer rows.Close()

	for rows.Next() {
		var eventName, tableName string
		var table *types.SQLTable

		err = rows.Scan(&tableName, &eventName)
		if err != nil {
			db.Log.InfoMsg("Error scanning table structure", "err", err)
			return tables, err
		}

		err = rows.Err()
		if err != nil {
			db.Log.InfoMsg("Error scanning table structure", "err", err)
			return tables, err
		}

		table, err = db.getTableDef(tableName)
		if err != nil {
			return tables, err
		}

		tables[tableName] = table
	}

	return tables, nil
}

// safe sanitizes a parameter
func safe(parameter string) string {
	replacer := strings.NewReplacer(";", "", ",", "")
	return replacer.Replace(parameter)
}

//getJSON returns marshaled json from JSON single column
func getJSON(JSON interface{}) ([]byte, error) {
	if JSON != nil {
		return json.Marshal(JSON)
	}
	return json.Marshal("")
}

//getJSONFromValues returns marshaled json from query values
func getJSONFromValues(values []interface{}) ([]byte, error) {
	if values != nil {
		return json.Marshal(values)
	}
	return json.Marshal("")
}

//getValuesFromJSON returns query values from unmarshaled JSON column
func getValuesFromJSON(JSON string) ([]interface{}, error) {
	pointers := make([]interface{}, 0)
	bytes := []byte(JSON)
	err := json.Unmarshal(bytes, &pointers)
	if err != nil {
		return nil, err
	}
	for i, ptr := range pointers {
		switch v := ptr.(type) {
		// Normalise integral floats
		case float64:
			i64 := int64(v)
			if float64(i64) == v {
				pointers[i] = i64
			}
		}
	}
	return pointers, nil
}

func digits(x uint64) int {
	if x == 0 {
		return 1
	}
	return int(math.Log10(float64(x))) + 1
}
