package types

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/hyperledger/burrow/event/query"
)

// EventSpec contains all event specifications
type EventSpec []EventDefinition

// EventDefinition struct (table name where to persist filtered events and it structure)
type EventDefinition struct {
	TableName    string
	Filter       string
	DeleteFilter string
	Columns      map[string]EventColumn
	query        query.Query
}

// Validate checks the structure of an EventDefinition
func (evDef EventDefinition) Validate() error {
	return validation.ValidateStruct(&evDef,
		validation.Field(&evDef.TableName, validation.Required, validation.Length(1, 60)),
		validation.Field(&evDef.Filter, validation.Required),
		validation.Field(&evDef.Columns, validation.Required, validation.Length(1, 0)),
	)
}

// Get a (memoised) Query from the EventDefinition Filter string
func (evDef EventDefinition) Query() (query.Query, error) {
	if evDef.query == nil {
		var err error
		evDef.query, err = query.New(evDef.Filter)
		if err != nil {
			return nil, err
		}
	}
	return evDef.query, nil
}

// EventColumn struct (table column definition)
type EventColumn struct {
	Name          string
	Type          string
	Primary       bool
	BytesToString bool
}

// Validate checks the structure of an EventColumn
func (evColumn EventColumn) Validate() error {
	return validation.ValidateStruct(&evColumn,
		validation.Field(&evColumn.Name, validation.Required, validation.Length(1, 60)),
	)
}
