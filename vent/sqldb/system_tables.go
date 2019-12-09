package sqldb

import (
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/vent/types"
)

// getSysTablesDefinition returns log, chain info & dictionary structures
func (db *SQLDB) systemTablesDefinition() types.EventTables {
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
					Name: columns.ChainID,
					Type: types.SQLColumnTypeVarchar,
				},
				{
					Name: columns.TimeStamp,
					Type: types.SQLColumnTypeTimeStamp,
				},
				{
					Name: columns.TableName,
					Type: types.SQLColumnTypeVarchar,
				},
				{
					Name: columns.EventName,
					Type: types.SQLColumnTypeVarchar,
				},
				{
					Name: columns.EventFilter,
					Type: types.SQLColumnTypeText,
				},
				// We use varchar for height - there is no uint64 type though numeric could have been used. We obtain the
				// maximum height by maxing over the serial ID type
				{
					Name: columns.Height,
					Type: types.SQLColumnTypeVarchar,
				},
				{
					Name:   columns.TxHash,
					Type:   types.SQLColumnTypeVarchar,
					Length: txs.HashLengthHex,
				},
				{
					Name:   columns.Action,
					Type:   types.SQLColumnTypeVarchar,
					Length: 50,
				},
				{
					Name: columns.DataRow,
					Type: types.SQLColumnTypeJSON,
				},
				{
					Name: columns.SqlStmt,
					Type: types.SQLColumnTypeText,
				},
				{
					Name: columns.SqlValues,
					Type: types.SQLColumnTypeText,
				},
			},
		},
		tables.Dictionary: {
			Name: tables.Dictionary,
			Columns: []*types.SQLTableColumn{
				{
					Name:    columns.TableName,
					Type:    types.SQLColumnTypeVarchar,
					Primary: true,
				},
				{
					Name:    columns.ColumnName,
					Type:    types.SQLColumnTypeVarchar,
					Primary: true,
				},
				{
					Name: columns.ColumnType,
					Type: types.SQLColumnTypeInt,
				},
				{
					Name: columns.ColumnLength,
					Type: types.SQLColumnTypeInt,
				},
				{
					Name: columns.PrimaryKey,
					Type: types.SQLColumnTypeInt,
				},
				{
					Name: columns.ColumnOrder,
					Type: types.SQLColumnTypeInt,
				},
			},
		},
		tables.ChainInfo: {
			Name: tables.ChainInfo,
			Columns: []*types.SQLTableColumn{
				{
					Name:    columns.ChainID,
					Type:    types.SQLColumnTypeVarchar,
					Primary: true,
				},
				{
					Name: columns.BurrowVersion,
					Type: types.SQLColumnTypeVarchar,
				},
				{
					Name: columns.Height,
					Type: types.SQLColumnTypeNumeric,
					// Maps to numeric(20, 0) - a 20 digit integer value
					Length: digits(maxUint64),
				},
			},
			NotifyChannels: map[string][]string{types.BlockHeightLabel: {columns.Height}},
		},
	}
}
