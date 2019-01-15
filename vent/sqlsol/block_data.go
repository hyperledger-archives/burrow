package sqlsol

import (
	"fmt"
	"strings"

	"github.com/hyperledger/burrow/vent/types"
)

// BlockData contains EventData definition
type BlockData struct {
	Data types.EventData
}

// NewBlockData returns a pointer to an empty BlockData structure
func NewBlockData() *BlockData {
	data := types.EventData{
		Block:  "",
		Tables: make(map[string]types.EventDataTable),
	}

	return &BlockData{
		Data: data,
	}
}

// GetBlockData returns the data structure
func (b *BlockData) GetBlockData() types.EventData {
	return b.Data
}

// GetBlockID gets block identification from structure
func (b *BlockData) GetBlockID() string {
	return b.Data.Block
}

// SetBlockID updates block identification in structure
func (b *BlockData) SetBlockID(blockID string) {
	b.Data.Block = blockID
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
func (b *BlockData) PendingRows(blockID string) bool {
	hasRows := false
	if strings.TrimSpace(b.Data.Block) == strings.TrimSpace(blockID) && len(b.Data.Tables) > 0 {
		hasRows = true
	}
	return hasRows
}
