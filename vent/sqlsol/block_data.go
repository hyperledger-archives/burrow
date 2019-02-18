package sqlsol

import (
	"fmt"

	"github.com/hyperledger/burrow/vent/types"
)

// BlockData contains EventData definition
type BlockData struct {
	Data types.EventData
}

// NewBlockData returns a pointer to an empty BlockData structure
func NewBlockData(height uint64) *BlockData {
	data := types.EventData{
		Tables:      make(map[string]types.EventDataTable),
		BlockHeight: height,
	}

	return &BlockData{
		Data: data,
	}
}

// AddRow appends a row to a specific table name in structure
func (b *BlockData) AddRow(tableName string, row types.EventDataRow) {
	if _, ok := b.Data.Tables[tableName]; !ok {
		b.Data.Tables[tableName] = types.EventDataTable{}
	}
	b.Data.Tables[tableName] = append(b.Data.Tables[tableName], row)
}

// GetRows gets data rows for a given table name from structure
func (b *BlockData) GetRows(tableName string) (types.EventDataTable, error) {
	if table, ok := b.Data.Tables[tableName]; ok {
		return table, nil
	}
	return nil, fmt.Errorf("GetRows: tableName does not exists as a table in data structure: %s ", tableName)
}

// PendingRows returns true if the given block has at least one pending row to upsert
func (b *BlockData) PendingRows(height uint64) bool {
	hasRows := false
	// TODO: understand why the guard on height is needed - what does it prevent?
	if b.Data.BlockHeight == height && len(b.Data.Tables) > 0 {
		hasRows = true
	}
	return hasRows
}
