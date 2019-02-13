package types

// DBAction generic type
type DBAction string

const (
	ActionDelete      DBAction = "DELETE"
	ActionUpsert      DBAction = "UPSERT"
	ActionRead        DBAction = "READ"
	ActionCreateTable DBAction = "CREATE"
	ActionAlterTable  DBAction = "ALTER"
	ActionInitialize  DBAction = "_INITIALIZE_VENT"
)

// EventData contains data for each block of events
// already mapped to SQL columns & tables
// Tables map key is the table name
type EventData struct {
	BlockHeight uint64
	Tables      map[string]EventDataTable
}

// EventDataTable is an array of rows
type EventDataTable []EventDataRow

// EventDataRow contains each SQL column name and a corresponding value to upsert
// map key is the column name and map value is the given column value
// if Action == 'delete' then the row has to be deleted
type EventDataRow struct {
	Action  DBAction
	RowData map[string]interface{}
	// The EventClass that caused this row to be emitted (if it was caused by an specific event)
	EventClass *EventClass
}
