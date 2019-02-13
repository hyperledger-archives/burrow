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
	tables := make(types.EventTables)
	blockCol := make(map[string]*types.SQLTableColumn)
	txCol := make(map[string]*types.SQLTableColumn)

	// block table
	blockCol[types.BlockHeightLabel] = &types.SQLTableColumn{
		Name:    types.SQLColumnLabelHeight,
		Type:    types.SQLColumnTypeVarchar,
		Length:  100,
		Primary: true,
		Order:   1,
	}

	blockCol[types.BlockHeaderLabel] = &types.SQLTableColumn{
		Name:    types.SQLColumnLabelBlockHeader,
		Type:    types.SQLColumnTypeJSON,
		Primary: false,
		Order:   2,
	}

	// transaction table
	txCol[types.BlockHeightLabel] = &types.SQLTableColumn{
		Name:    types.SQLColumnLabelHeight,
		Type:    types.SQLColumnTypeVarchar,
		Length:  100,
		Primary: true,
		Order:   1,
	}

	txCol[types.TxTxHashLabel] = &types.SQLTableColumn{
		Name:    types.SQLColumnLabelTxHash,
		Type:    types.SQLColumnTypeVarchar,
		Length:  txs.HashLengthHex,
		Primary: true,
		Order:   2,
	}

	txCol[types.TxIndexLabel] = &types.SQLTableColumn{
		Name:    types.SQLColumnLabelIndex,
		Type:    types.SQLColumnTypeNumeric,
		Length:  0,
		Primary: false,
		Order:   3,
	}

	txCol[types.TxTxTypeLabel] = &types.SQLTableColumn{
		Name:    types.SQLColumnLabelTxType,
		Type:    types.SQLColumnTypeVarchar,
		Length:  100,
		Primary: false,
		Order:   4,
	}

	txCol[types.TxEnvelopeLabel] = &types.SQLTableColumn{
		Name:    types.SQLColumnLabelEnvelope,
		Type:    types.SQLColumnTypeJSON,
		Primary: false,
		Order:   5,
	}

	txCol[types.TxEventsLabel] = &types.SQLTableColumn{
		Name:    types.SQLColumnLabelEvents,
		Type:    types.SQLColumnTypeJSON,
		Primary: false,
		Order:   6,
	}

	txCol[types.TxResultLabel] = &types.SQLTableColumn{
		Name:    types.SQLColumnLabelResult,
		Type:    types.SQLColumnTypeJSON,
		Primary: false,
		Order:   7,
	}

	txCol[types.TxReceiptLabel] = &types.SQLTableColumn{
		Name:    types.SQLColumnLabelReceipt,
		Type:    types.SQLColumnTypeJSON,
		Primary: false,
		Order:   8,
	}

	txCol[types.TxExceptionLabel] = &types.SQLTableColumn{
		Name:    types.SQLColumnLabelException,
		Type:    types.SQLColumnTypeJSON,
		Primary: false,
		Order:   9,
	}

	// add tables
	tables[types.SQLBlockTableName] = types.SQLTable{
		Name:    types.SQLBlockTableName,
		Columns: blockCol,
	}

	tables[types.SQLTxTableName] = types.SQLTable{
		Name:    types.SQLTxTableName,
		Columns: txCol,
	}

	return tables
}
