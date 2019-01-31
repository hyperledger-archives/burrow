package types

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/hyperledger/burrow/event/query"
)

// EventSpec contains all event specifications
type EventSpec []EventDefinition

// EventDefinition struct (table name where to persist filtered events and it structure)
type EventDefinition struct {
	TableName    string                 `json:"TableName"`
	Filter       string                 `json:"Filter"`
	DeleteFilter string                 `json:"DeleteFilter"`
	Columns      map[string]EventColumn `json:"Columns"`
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
	Name          string `json:"name"`
	Type          string `json:"type"`
	Primary       bool   `json:"primary"`
	BytesToString bool   `json:"bytesToString"`
}

// Validate checks the structure of an EventColumn
func (evColumn EventColumn) Validate() error {
	return validation.ValidateStruct(&evColumn,
		validation.Field(&evColumn.Name, validation.Required, validation.Length(1, 60)),
	)
}
