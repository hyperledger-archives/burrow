package types

import (
	"github.com/alecthomas/jsonschema"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/hyperledger/burrow/event/query"
)

// EventSpec contains all event class specifications
type EventSpec []EventClass

func EventSpecSchema() *jsonschema.Schema {
	return jsonschema.Reflect(EventSpec{})
}

// EventClass struct (table name where to persist filtered events and it structure)
type EventClass struct {
	// Destination table in DB
	TableName string
	// Burrow event filter query in query peg grammar
	Filter string
	// The name of a solidity event field that when present indicates that the rest of the event should be interpreted
	// as requesting a row deletion (rather than upsert) in the projection table.
	DeleteMarkerField string `json:",omitempty"`
	// Mapping from solidity event field name to EventField descriptor
	Fields map[string]EventField
	query  query.Query
}

// Validate checks the structure of an EventClass
func (ec EventClass) Validate() error {
	return validation.ValidateStruct(&ec,
		validation.Field(&ec.TableName, validation.Required, validation.Length(1, 60)),
		validation.Field(&ec.Filter, validation.Required),
		validation.Field(&ec.Fields, validation.Required, validation.Length(1, 0)),
	)
}

// Get a (memoised) Query from the EventClass Filter string
func (ec EventClass) Query() (query.Query, error) {
	if ec.query == nil {
		var err error
		ec.query, err = query.New(ec.Filter)
		if err != nil {
			return nil, err
		}
	}
	return ec.query, nil
}

// EventField struct (table column definition)
type EventField struct {
	// SQL column name to which to map this event field
	ColumnName string
	// EVM type of this field
	Type string
	// Whether this event field should map to a primary key
	Primary bool `json:",omitempty"`
	// Whether to convert this event field from bytes32 to string
	BytesToString bool `json:",omitempty"`
	// Notification channels on which submit (via a trigger) a payload that contains this column's new value (upsert) or
	// old value (delete). The payload will contain all other values with the same channel set as a JSON object.
	Notify []string `json:",omitempty"`
}

// Validate checks the structure of an EventField
func (evColumn EventField) Validate() error {
	return validation.ValidateStruct(&evColumn,
		validation.Field(&evColumn.ColumnName, validation.Required, validation.Length(1, 60)),
	)
}
