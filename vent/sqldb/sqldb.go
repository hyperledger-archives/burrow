package sqldb

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/burrow/vent/logger"
	"github.com/hyperledger/burrow/vent/sqldb/adapters"
	"github.com/hyperledger/burrow/vent/types"
)

// SQLDB implements the access to a sql database
type SQLDB struct {
	DB        *sql.DB
	DBAdapter adapters.DBAdapter
	Schema    string
	Log       *logger.Logger
}

// NewSQLDB delegates work to a specific database adapter implementation,
// opens database connection and create log tables
func NewSQLDB(connection types.SQLConnection) (*SQLDB, error) {
	db := &SQLDB{
		Schema: connection.DBSchema,
		Log:    connection.Log,
	}

	var url string

	switch connection.DBAdapter {
	case types.PostgresDB:
		db.DBAdapter = adapters.NewPostgresAdapter(safe(connection.DBSchema), connection.Log)
		url = connection.DBURL

	case types.SQLiteDB:
		db.DBAdapter = adapters.NewSQLiteAdapter(connection.Log)
		// "?_journal_mode=WAL" parameter is necessary to prevent database locking
		url = connection.DBURL + "?_journal_mode=WAL"

	default:
		return nil, errors.New("invalid database adapter")
	}

	var err error
	db.DB, err = db.DBAdapter.Open(url)
	if err != nil {
		db.Log.Info("msg", "Error opening database connection", "err", err)
		return nil, err
	}

	if err = db.Ping(); err != nil {
		db.Log.Info("msg", "Error database not available", "err", err)
		return nil, err
	}

	db.Log.Info("msg", "Initializing DB")

	// Create dictionary and log tables
	sysTables := db.getSysTablesDefinition()

	// IMPORTANT: DO NOT CHANGE TABLE CREATION ORDER (1)
	if err = db.createTable(sysTables[types.SQLDictionaryTableName], string(types.ActionInitialize)); err != nil {
		if !db.DBAdapter.ErrorEquals(err, types.SQLErrorTypeDuplicatedTable) {
			db.Log.Info("msg", "Error creating Dictionary table", "err", err)
			return nil, err
		}
	}

	// IMPORTANT: DO NOT CHANGE TABLE CREATION ORDER (2)
	if err = db.createTable(sysTables[types.SQLLogTableName], string(types.ActionInitialize)); err != nil {
		if !db.DBAdapter.ErrorEquals(err, types.SQLErrorTypeDuplicatedTable) {
			db.Log.Info("msg", "Error creating Log table", "err", err)
			return nil, err
		}
	}

	// IMPORTANT: DO NOT CHANGE TABLE CREATION ORDER (3)
	if err = db.createTable(sysTables[types.SQLChainInfoTableName], string(types.ActionInitialize)); err != nil {
		if !db.DBAdapter.ErrorEquals(err, types.SQLErrorTypeDuplicatedTable) {
			db.Log.Info("msg", "Error creating Chain Info table", "err", err)
			return nil, err
		}
	}

	if err = db.CleanTables(connection.ChainID, connection.BurrowVersion); err != nil {
		db.Log.Info("msg", "Error cleaning tables", "err", err)
		return nil, err
	}
	return db, nil
}

// CleanTables, drop tables if stored chainID is different from the given one & store new chainID
// if the chainID is the same, do nothing
func (db *SQLDB) CleanTables(chainID, burrowVersion string) error {

	if chainID == "" {
		return fmt.Errorf("error CHAIN ID cannot by empty")
	}

	cleanQueries := db.DBAdapter.CleanDBQueries()

	var savedChainID, savedBurrowVersion, query string
	savedRows := 0

	// Read chainID
	query = cleanQueries.SelectChainIDQry
	if err := db.DB.QueryRow(query).Scan(&savedRows, &savedChainID, &savedBurrowVersion); err != nil {
		db.Log.Info("msg", "Error selecting CHAIN ID", "err", err, "query", query)
		return err
	}

	switch {
	// Must be empty or one row
	case savedRows != 0 && savedRows != 1:
		return fmt.Errorf("error multiple CHAIN ID returned")

	// First database access
	case savedRows == 0:
		// Save new values and exit
		query = cleanQueries.InsertChainIDQry
		if _, err := db.DB.Exec(query, chainID, burrowVersion); err != nil {
			db.Log.Info("msg", "Error inserting CHAIN ID", "err", err, "query", query)
			return err
		}
		return nil

	// if data equals previous version exit
	case savedChainID == chainID:
		return nil

	// clean database
	default:
		var tx *sql.Tx
		var err error
		var tableName string
		tables := make([]string, 0)

		// Begin tx
		if tx, err = db.DB.Begin(); err != nil {
			db.Log.Info("msg", "Error beginning transaction", "err", err)
			return err
		}
		defer tx.Rollback()

		// Delete chainID
		query := cleanQueries.DeleteChainIDQry
		if _, err = tx.Exec(query); err != nil {
			db.Log.Info("msg", "Error deleting CHAIN ID", "err", err, "query", query)
			return err
		}

		// Insert chainID
		query = cleanQueries.InsertChainIDQry
		if _, err := tx.Exec(query, chainID, burrowVersion); err != nil {
			db.Log.Info("msg", "Error inserting CHAIN ID", "err", err, "query", query)
			return err
		}

		// Load Tables
		query = cleanQueries.SelectDictionaryQry
		rows, err := tx.Query(query)
		if err != nil {
			db.Log.Info("msg", "error querying dictionary", "err", err, "query", query)
			return err
		}
		defer rows.Close()

		for rows.Next() {

			if err = rows.Scan(&tableName); err != nil {
				db.Log.Info("msg", "error scanning table structure", "err", err)
				return err
			}

			if err = rows.Err(); err != nil {
				db.Log.Info("msg", "error scanning table structure", "err", err)
				return err
			}
			tables = append(tables, tableName)
		}

		// Delete Dictionary
		query = cleanQueries.DeleteDictionaryQry
		if _, err = tx.Exec(query); err != nil {
			db.Log.Info("msg", "Error deleting dictionary", "err", err, "query", query)
			return err
		}

		// Delete Log
		query = cleanQueries.DeleteLogQry
		if _, err = tx.Exec(query); err != nil {
			db.Log.Info("msg", "Error deleting log", "err", err, "query", query)
			return err
		}
		// Drop database tables
		for _, tableName = range tables {
			query = db.DBAdapter.DropTableQuery(tableName)
			if _, err = tx.Exec(query); err != nil {
				// if error == table does not exists, continue
				if !db.DBAdapter.ErrorEquals(err, types.SQLErrorTypeUndefinedTable) {
					db.Log.Info("msg", "error dropping tables", "err", err, "value", tableName, "query", query)
					return err
				}
			}
		}

		// Commit
		if err = tx.Commit(); err != nil {
			db.Log.Info("msg", "Error commiting transaction", "err", err)
			return err
		}

		return nil
	}
}

// Close database connection
func (db *SQLDB) Close() {
	if err := db.DB.Close(); err != nil {
		db.Log.Error("msg", "Error closing database", "err", err)
	}
}

// Ping database
func (db *SQLDB) Ping() error {
	if err := db.DB.Ping(); err != nil {
		db.Log.Info("msg", "Error database not available", "err", err)
		return err
	}

	return nil
}

// GetLastBlockID returns last inserted blockId from log table
func (db *SQLDB) GetLastBlockHeight() (uint64, error) {
	query := db.DBAdapter.LastBlockIDQuery()
	id := ""

	db.Log.Info("msg", "MAX ID", "query", query)

	if err := db.DB.QueryRow(query).Scan(&id); err != nil {
		db.Log.Info("msg", "Error selecting last block id", "err", err)
		return 0, err
	}
	height, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse height from block ID: %v", err)
	}
	return height, nil
}

// SynchronizeDB synchronize db tables structures from given tables specifications
func (db *SQLDB) SynchronizeDB(eventTables types.EventTables) error {
	db.Log.Info("msg", "Synchronizing DB")

	for eventName, table := range eventTables {
		found, err := db.findTable(table.Name)
		if err != nil {
			return err
		}

		if found {
			err = db.alterTable(table, eventName)
		} else {
			err = db.createTable(table, eventName)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// SetBlock inserts or updates multiple rows and stores log info in SQL tables
func (db *SQLDB) SetBlock(eventTables types.EventTables, eventData types.EventData) error {
	db.Log.Info("msg", "Synchronize Block..........")

	//Declarations
	var logStmt *sql.Stmt
	var tx *sql.Tx
	var safeTable string
	var query string
	var queryVal types.UpsertDeleteQuery
	var err error
	var errQuery error
	var txHash interface{}
	var jsonData []byte
	var sqlValues []byte

	// Begin tx
	if tx, err = db.DB.Begin(); err != nil {
		db.Log.Info("msg", "Error beginning transaction", "err", err)
		return err
	}
	defer tx.Rollback()

	// Prepare log statement
	logQuery := db.DBAdapter.InsertLogQuery()
	if logStmt, err = tx.Prepare(logQuery); err != nil {
		db.Log.Info("msg", "Error preparing log stmt", "err", err)
		return err
	}

loop:
	// for each table in the block
	for eventName, table := range eventTables {
		safeTable = safe(table.Name)
		dataRows := eventData.Tables[table.Name]
		// for Each Row
		for _, row := range dataRows {

			switch row.Action {
			case types.ActionUpsert:
				//Prepare Upsert
				if queryVal, txHash, errQuery = db.DBAdapter.UpsertQuery(table, row); errQuery != nil {
					db.Log.Info("msg", "Error building upsert query", "err", errQuery, "value", fmt.Sprintf("%v %v", table, row))
					break loop // exits from all loops -> continue in close log stmt
				}

			case types.ActionDelete:
				//Prepare Delete
				txHash = nil
				if queryVal, errQuery = db.DBAdapter.DeleteQuery(table, row); errQuery != nil {
					db.Log.Info("msg", "Error building delete query", "err", errQuery, "value", fmt.Sprintf("%v %v", table, row))
					break loop // exits from all loops -> continue in close log stmt
				}
			default:
				//Invalid Action
				db.Log.Info("msg", "invalid action", "value", row.Action)
				err = fmt.Errorf("invalid row action %s", row.Action)
				break loop // exits from all loops -> continue in close log stmt
			}

			query = queryVal.Query

			// Perform row action
			db.Log.Info("msg", row.Action, "query", query, "value", queryVal.Values)
			if _, err = tx.Exec(query, queryVal.Pointers...); err != nil {
				db.Log.Info("msg", fmt.Sprintf("error performing %s on row", row.Action), "err", err, "value", queryVal.Values)
				break loop // exits from all loops -> continue in close log stmt
			}

			// Marshal the rowData map
			if jsonData, err = db.getJSON(row.RowData); err != nil {
				db.Log.Info("msg", "error marshaling rowData", "err", err, "value", fmt.Sprintf("%v", row.RowData))
				break loop // exits from all loops -> continue in close log stmt
			}

			// Marshal sql values
			if sqlValues, err = db.getJSONFromValues(queryVal.Pointers); err != nil {
				db.Log.Info("msg", "error marshaling rowdata", "err", err, "value", fmt.Sprintf("%v", row.RowData))
				break loop // exits from all loops -> continue in close log stmt
			}

			// Insert in log
			db.Log.Info("msg", "INSERT LOG", "query", logQuery, "value",
				fmt.Sprintf("tableName = %s eventName = %s block = %d", safeTable, eventName, eventData.BlockHeight))

			if _, err = logStmt.Exec(safeTable, eventName, row.EventClass.GetFilter(), eventData.BlockHeight, txHash,
				row.Action, jsonData, query, sqlValues); err != nil {
				db.Log.Info("msg", "Error inserting into log", "err", err)
				break loop // exits from all loops -> continue in close log stmt
			}
		}
	}

	// Close log statement
	if err == nil {
		if err = logStmt.Close(); err != nil {
			db.Log.Info("msg", "Error closing log stmt", "err", err)
		}
	}

	// Error handling
	if err != nil {
		// Rollback error
		if errRb := tx.Rollback(); errRb != nil {
			db.Log.Info("msg", "Error on rollback", "err", errRb)
			return errRb
		}

		//Is a SQL error
		if db.DBAdapter.ErrorEquals(err, types.SQLErrorTypeGeneric) {

			// Table does not exists
			if db.DBAdapter.ErrorEquals(err, types.SQLErrorTypeUndefinedTable) {
				db.Log.Warn("msg", "Table not found", "value", safeTable)
				//Synchronize DB
				if err = db.SynchronizeDB(eventTables); err != nil {
					return err
				}
				//Retry
				return db.SetBlock(eventTables, eventData)
			}

			// Columns do not match
			if db.DBAdapter.ErrorEquals(err, types.SQLErrorTypeUndefinedColumn) {
				db.Log.Warn("msg", "Column not found", "value", safeTable)
				//Synchronize DB
				if err = db.SynchronizeDB(eventTables); err != nil {
					return err
				}
				//Retry
				return db.SetBlock(eventTables, eventData)
			}
			return err
		}
		return err
	}

	db.Log.Info("msg", "COMMIT")

	if err := tx.Commit(); err != nil {
		db.Log.Info("msg", "Error on commit", "err", err)
		return err
	}

	return nil
}

// GetBlock returns all tables structures and row data for given block
func (db *SQLDB) GetBlock(height uint64) (types.EventData, error) {
	var data types.EventData
	data.BlockHeight = height
	data.Tables = make(map[string]types.EventDataTable)

	// get all table structures involved in the block
	tables, err := db.getBlockTables(height)
	if err != nil {
		return data, err
	}

	query := ""

	// for each table
	for _, table := range tables {
		// get query for table
		query, err = db.getSelectQuery(table, height)
		if err != nil {
			db.Log.Info("msg", "Error building table query", "err", err)
			return data, err
		}
		query = query
		db.Log.Info("msg", "Query table data", "query", query)
		rows, err := db.DB.Query(query)
		if err != nil {
			db.Log.Info("msg", "Error querying table data", "err", err)
			return data, err
		}
		defer rows.Close()

		cols, err := rows.Columns()
		if err != nil {
			db.Log.Info("msg", "Error getting row columns", "err", err)
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
				db.Log.Info("msg", "Error scanning data", "err", err)
				return data, err
			}
			db.Log.Info("msg", "Query resultset", "value", fmt.Sprintf("%+v", containers))

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
			db.Log.Info("msg", "Error during rows iteration", "err", err)
			return data, err
		}
		data.Tables[table.Name] = dataRows
	}
	return data, nil
}

// RestoreDB restores the DB to a given moment in time
func (db *SQLDB) RestoreDB(time time.Time, prefix string) error {

	const yymmddhhmmss = "2006-01-02 15:04:05"

	var pointers []interface{}

	if prefix == "" {
		return fmt.Errorf("error prefix mus not be empty")
	}

	// Get Restore DB query
	query := db.DBAdapter.RestoreDBQuery()
	strTime := time.Format(yymmddhhmmss)

	db.Log.Info("msg", "RESTORING DB..................................")

	// Open rows
	db.Log.Info("msg", "open log", "query", query)
	rows, err := db.DB.Query(query, strTime)
	if err != nil {
		db.Log.Info("msg", "error querying log", "err", err)
		return err
	}
	defer rows.Close()

	// For each row returned
	for rows.Next() {
		var tableName, sqlSmt, sqlValues string
		var action types.DBAction

		if err = rows.Scan(&tableName, &action, &sqlSmt, &sqlValues); err != nil {
			db.Log.Info("msg", "error scanning table structure", "err", err)
			return err
		}

		if err = rows.Err(); err != nil {
			db.Log.Info("msg", "error scanning table structure", "err", err)
			return err
		}

		restoreTable := fmt.Sprintf("%s_%s", prefix, tableName)

		switch action {
		case types.ActionUpsert, types.ActionDelete:
			// get row values
			if pointers, err = db.getValuesFromJSON(sqlValues); err != nil {
				db.Log.Info("msg", "error unmarshaling json", "err", err, "value", sqlValues)
				return err
			}

			// Prepare Upsert/delete
			query = strings.Replace(sqlSmt, tableName, restoreTable, -1)

			db.Log.Info("msg", "SQL COMMAND", "sql", query)
			if _, err = db.DB.Exec(query, pointers...); err != nil {
				db.Log.Info("msg", "Error executing upsert/delete ", "err", err, "value", sqlSmt, "data", sqlValues)
				return err
			}

		case types.ActionAlterTable, types.ActionCreateTable:
			// Prepare Alter/Create Table
			query = strings.Replace(sqlSmt, tableName, restoreTable, -1)

			db.Log.Info("msg", "SQL COMMAND", "sql", query)
			if _, err = db.DB.Exec(query); err != nil {
				db.Log.Info("msg", "Error executing alter/create table command ", "err", err, "value", sqlSmt)
				return err
			}
		default:
			// Invalid Action
			db.Log.Info("msg", "invalid action", "value", action)
			return fmt.Errorf("invalid row action %s", action)
		}
	}
	return nil
}
