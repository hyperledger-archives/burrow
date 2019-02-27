package sqlsol

import (
	"fmt"

	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/vent/types"
)

// SpecLoader loads spec files and parses them
func SpecLoader(specFileOrDir string, createBlkTxTables bool) (*Projection, error) {
	var projection *Projection
	var err error

	if specFileOrDir == "" {
		return nil, fmt.Errorf("please provide a spec file or directory")
	}

	projection, err = NewProjectionFromFolder(specFileOrDir)
	if err != nil {
		return nil, fmt.Errorf("error parsing spec: %v", err)
	}

	if createBlkTxTables {
		// add block & tx to tables definition
		blkTxTables := getBlockTxTablesDefinition()

		for k, v := range blkTxTables {
			projection.Tables[k] = v
		}

	}

	return projection, nil
}

// getBlockTxTablesDefinition returns block & transaction structures
func getBlockTxTablesDefinition() types.EventTables {
	return types.EventTables{
		types.SQLBlockTableName: &types.SQLTable{
			Name: types.SQLBlockTableName,
			Columns: []*types.SQLTableColumn{
				{
					Name:    types.SQLColumnLabelHeight,
					Type:    types.SQLColumnTypeVarchar,
					Length:  100,
					Primary: true,
				},
				{
					Name:    types.SQLColumnLabelBlockHeader,
					Type:    types.SQLColumnTypeJSON,
					Primary: false,
				},
			},
		},

		types.SQLTxTableName: &types.SQLTable{
			Name: types.SQLTxTableName,
			Columns: []*types.SQLTableColumn{
				// transaction table
				{
					Name:    types.SQLColumnLabelHeight,
					Type:    types.SQLColumnTypeVarchar,
					Length:  100,
					Primary: true,
				},
				{
					Name:    types.SQLColumnLabelTxHash,
					Type:    types.SQLColumnTypeVarchar,
					Length:  txs.HashLengthHex,
					Primary: true,
				},
				{
					Name:    types.SQLColumnLabelIndex,
					Type:    types.SQLColumnTypeNumeric,
					Length:  0,
					Primary: false,
				},
				{
					Name:    types.SQLColumnLabelTxType,
					Type:    types.SQLColumnTypeVarchar,
					Length:  100,
					Primary: false,
				},
				{
					Name:    types.SQLColumnLabelEnvelope,
					Type:    types.SQLColumnTypeJSON,
					Primary: false,
				},
				{
					Name:    types.SQLColumnLabelEvents,
					Type:    types.SQLColumnTypeJSON,
					Primary: false,
				},
				{
					Name:    types.SQLColumnLabelResult,
					Type:    types.SQLColumnTypeJSON,
					Primary: false,
				},
				{
					Name:    types.SQLColumnLabelReceipt,
					Type:    types.SQLColumnTypeJSON,
					Primary: false,
				},
				{
					Name:    types.SQLColumnLabelException,
					Type:    types.SQLColumnTypeJSON,
					Primary: false,
				},
			},
		},
	}
}
