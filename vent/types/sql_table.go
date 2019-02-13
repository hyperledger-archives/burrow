package types

import (
	"fmt"
)

// SQLTable contains the structure of a SQL table,
type SQLTable struct {
	Name    string
	Columns map[string]*SQLTableColumn
	// Map of channel name -> columns to be sent as payload on that channel
	NotifyChannels map[string][]string
}

// SQLTableColumn contains the definition of a SQL table column,
// the Order is given to be able to sort the columns to be created
type SQLTableColumn struct {
	Name    string
	Type    SQLColumnType
	Primary bool
	Length  int
	Order   int
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
	return fmt.Sprintf("SQLTableColumn{%02d-%s%s: %v%s}",
		col.Order, col.Name, primaryString, col.Type, lengthString)
}

// UpsertDeleteQuery contains query and values to upsert or delete row data
type UpsertDeleteQuery struct {
	Query    string
	Values   string
	Pointers []interface{}
}

// SQL log & dictionary tables
const (
	SQLLogTableName        = "_vent_log"
	SQLDictionaryTableName = "_vent_dictionary"
	SQLBlockTableName      = "_vent_block"
	SQLTxTableName         = "_vent_tx"
	SQLChainInfoTableName  = "_vent_chain"
)

// fixed sql column names in tables
const (
	// log
	SQLColumnLabelId          = "_id"
	SQLColumnLabelTimeStamp   = "_timestamp"
	SQLColumnLabelTableName   = "_tablename"
	SQLColumnLabelEventName   = "_eventname"
	SQLColumnLabelEventFilter = "_eventfilter"
	SQLColumnLabelHeight      = "_height"
	SQLColumnLabelTxHash      = "_txhash"
	SQLColumnLabelAction      = "_action"
	SQLColumnLabelDataRow     = "_datarow"
	SQLColumnLabelSqlStmt     = "_sqlstmt"
	SQLColumnLabelSqlValues   = "_sqlvalues"

	// dictionary
	SQLColumnLabelColumnName   = "_columnname"
	SQLColumnLabelColumnType   = "_columntype"
	SQLColumnLabelColumnLength = "_columnlength"
	SQLColumnLabelPrimaryKey   = "_primarykey"
	SQLColumnLabelColumnOrder  = "_columnorder"

	// chain info
	SQLColumnLabelBurrowVer = "_burrowversion"
	SQLColumnLabelChainID   = "_chainid"

	// context
	SQLColumnLabelIndex       = "_index"
	SQLColumnLabelEventType   = "_eventtype"
	SQLColumnLabelBlockHeader = "_blockheader"
	SQLColumnLabelTxType      = "_txtype"
	SQLColumnLabelEnvelope    = "_envelope"
	SQLColumnLabelEvents      = "_events"
	SQLColumnLabelResult      = "_result"
	SQLColumnLabelReceipt     = "_receipt"
	SQLColumnLabelException   = "_exception"
)

// labels for column mapping
const (
	// event related
	EventNameLabel = "eventName"
	EventTypeLabel = "eventType"

	// block related
	BlockHeightLabel = "height"
	BlockHeaderLabel = "blockHeader"

	// transaction related
	TxTxTypeLabel    = "txType"
	TxTxHashLabel    = "txHash"
	TxIndexLabel     = "index"
	TxEnvelopeLabel  = "envelope"
	TxEventsLabel    = "events"
	TxResultLabel    = "result"
	TxReceiptLabel   = "receipt"
	TxExceptionLabel = "exception"
)
