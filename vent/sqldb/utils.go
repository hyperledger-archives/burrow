package sqldb

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/hyperledger/burrow/vent/sqldb/adapters"

	"github.com/hyperledger/burrow/txs"

	"encoding/json"

	"github.com/hyperledger/burrow/vent/types"
)

const maxUint64 uint64 = (1 << 64) - 1

var tables = types.DefaultSQLTableNames
var columns = types.DefaultSQLColumnNames

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

// getSysTablesDefinition returns log, chain info & dictionary structures
func (db *SQLDB) getSysTablesDefinition() types.EventTables {
	return types.EventTables{
		tables.Log: {
			Name: tables.Log,
			Columns: []*types.SQLTableColumn{
				{
					Name:    columns.Id,
					Type:    types.SQLColumnTypeSerial,
					Primary: true,
				},
				{
					Name:   columns.ChainID,
					Type:   types.SQLColumnTypeVarchar,
					Length: 100,
				},
				{
					Name: columns.TimeStamp,
					Type: types.SQLColumnTypeTimeStamp,
				},
				{
					Name:   columns.TableName,
					Type:   types.SQLColumnTypeVarchar,
					Length: 100,
				},
				{
					Name:   columns.EventName,
					Type:   types.SQLColumnTypeVarchar,
					Length: 100,
				},
				{
					Name:   columns.EventFilter,
					Type:   types.SQLColumnTypeVarchar,
					Length: 100,
				},
				// We use varchar for height - there is no uint64 type though numeric could have been used. We obtain the
				// maximum height by maxing over the serial ID type
				{
					Name:   columns.Height,
					Type:   types.SQLColumnTypeVarchar,
					Length: 100,
				},
				{
					Name:   columns.TxHash,
					Type:   types.SQLColumnTypeVarchar,
					Length: txs.HashLengthHex,
				},
				{
					Name:   columns.Action,
					Type:   types.SQLColumnTypeVarchar,
					Length: 20,
				},
				{
					Name:   columns.DataRow,
					Type:   types.SQLColumnTypeJSON,
					Length: 0,
				},
				{
					Name:   columns.SqlStmt,
					Type:   types.SQLColumnTypeText,
					Length: 0,
				},
				{
					Name:   columns.SqlValues,
					Type:   types.SQLColumnTypeText,
					Length: 0,
				},
			},
			NotifyChannels: map[string][]string{types.BlockHeightLabel: {columns.Height}},
		},
		tables.Dictionary: {
			Name: tables.Dictionary,
			Columns: []*types.SQLTableColumn{
				{
					Name:    columns.TableName,
					Type:    types.SQLColumnTypeVarchar,
					Length:  100,
					Primary: true,
				},
				{
					Name:    columns.ColumnName,
					Type:    types.SQLColumnTypeVarchar,
					Length:  100,
					Primary: true,
				},
				{
					Name:   columns.ColumnType,
					Type:   types.SQLColumnTypeInt,
					Length: 0,
				},
				{
					Name:   columns.ColumnLength,
					Type:   types.SQLColumnTypeInt,
					Length: 0,
				},
				{
					Name:   columns.PrimaryKey,
					Type:   types.SQLColumnTypeInt,
					Length: 0,
				},
				{
					Name:   columns.ColumnOrder,
					Type:   types.SQLColumnTypeInt,
					Length: 0,
				},
			},
		},
		tables.ChainInfo: {
			Name: tables.ChainInfo,
			Columns: []*types.SQLTableColumn{
				{
					Name:    columns.ChainID,
					Type:    types.SQLColumnTypeVarchar,
					Length:  100,
					Primary: true,
				},
				{
					Name:   columns.BurrowVersion,
					Type:   types.SQLColumnTypeVarchar,
					Length: 100,
				},
				{
					Name: columns.Height,
					Type: types.SQLColumnTypeNumeric,
					// Maps to numeric(20, 0) - a 20 digit integer value
					Length: digits(maxUint64),
				},
			},
		},
	}
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
