package types

import (
	"fmt"
)

// SQLTable contains the structure of a SQL table,
type SQLTable struct {
	Name    string
	Columns []*SQLTableColumn
	// Map of channel name -> columns to be sent as payload on that channel
	NotifyChannels map[string][]string
	columns        map[string]*SQLTableColumn
}

func (table *SQLTable) GetColumn(columnName string) *SQLTableColumn {
	if table.columns == nil {
		table.columns = make(map[string]*SQLTableColumn, len(table.Columns))
		for _, column := range table.Columns {
			table.columns[column.Name] = column
		}
	}
	return table.columns[columnName]
}

// SQLTableColumn contains the definition of a SQL table column,
// the Order is given to be able to sort the columns to be created
type SQLTableColumn struct {
	Name    string
	Type    SQLColumnType
	Primary bool
	Length  int
}

func (col *SQLTableColumn) String() string {
	primaryString := ""
	if col.Primary {
		primaryString = " (primary)"
	}
	lengthString := ""
	if col.Length != 0 {
		lengthString = fmt.Sprintf(" (length %d)", col.Length)
	}
	return fmt.Sprintf("SQLTableColumn{%s%s: %v%s}",
		col.Name, primaryString, col.Type, lengthString)
}

func (col *SQLTableColumn) Equals(otherCol *SQLTableColumn) bool {
	columnA := *col
	columnB := *otherCol
	return columnA == columnB
}

// UpsertDeleteQuery contains query and values to upsert or delete row data
type UpsertDeleteQuery struct {
	Query    string
	Values   string
	Pointers []interface{}
}

type SQLNames struct {
	Tables  SQLTableNames
	Columns SQLColumnNames
}

var DefaultSQLNames = SQLNames{
	Tables:  DefaultSQLTableNames,
	Columns: DefaultSQLColumnNames,
}

type SQLTableNames struct {
	Log        string
	Dictionary string
	Block      string
	Tx         string
	ChainInfo  string
}

var DefaultSQLTableNames = SQLTableNames{
	Log:        "_vent_log",
	Dictionary: "_vent_dictionary",
	Block:      "_vent_block",
	Tx:         "_vent_tx",
	ChainInfo:  "_vent_chain",
}

type SQLColumnNames struct {
	// log
	Id          string
	TimeStamp   string
	TableName   string
	EventName   string
	EventFilter string
	Height      string
	TxHash      string
	Action      string
	DataRow     string
	SqlStmt     string
	SqlValues   string
	// dictionary
	ColumnName   string
	ColumnType   string
	ColumnLength string
	PrimaryKey   string
	ColumnOrder  string
	// chain info
	BurrowVersion string
	ChainID       string
	// context
	Index       string
	EventIndex  string
	EventType   string
	BlockHeader string
	TxType      string
	Envelope    string
	Events      string
	Result      string
	Receipt     string
	Origin      string
	Exception   string
}

var DefaultSQLColumnNames = SQLColumnNames{
	// log
	Id:          "_id",
	TimeStamp:   "_timestamp",
	TableName:   "_tablename",
	EventName:   "_eventname",
	EventFilter: "_eventfilter",
	Height:      "_height",
	TxHash:      "_txhash",
	Action:      "_action",
	DataRow:     "_datarow",
	SqlStmt:     "_sqlstmt",
	SqlValues:   "_sqlvalues",
	// dictionary,
	ColumnName:   "_columnname",
	ColumnType:   "_columntype",
	ColumnLength: "_columnlength",
	PrimaryKey:   "_primarykey",
	ColumnOrder:  "_columnorder",
	// chain info,
	BurrowVersion: "_burrowversion",
	ChainID:       "_chainid",
	// context,
	Index:       "_index",
	EventIndex:  "_eventindex",
	EventType:   "_eventtype",
	BlockHeader: "_blockheader",
	TxType:      "_txtype",
	Envelope:    "_envelope",
	Events:      "_events",
	Result:      "_result",
	Receipt:     "_receipt",
	Origin:      "_origin",
	Exception:   "_exception",
}

// labels for column mapping
const (
	// event related
	EventNameLabel  = "eventName"
	EventTypeLabel  = "eventType"
	EventIndexLabel = "eventIndex"

	// block related
	ChainIDLabel     = "chainid"
	BlockHeightLabel = "height"
	BlockIndexLabel  = "index"

	// transaction related
	TxTxHashLabel = "txHash"
)
