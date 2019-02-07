package sqlsol

import (
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/vent/types"
	"github.com/pkg/errors"
)

// SpecLoader loads spec files and parses them
func SpecLoader(specDir, specFile string, createBlkTxTables bool) (*Projection, error) {

	var projection *Projection
	var err error

	if specDir == "" && specFile == "" {
		return nil, errors.New("One of SpecDir or SpecFile must be provided")
	}

	if specDir != "" && specFile != "" {
		return nil, errors.New("SpecDir or SpecFile must be provided, but not both")
	}

	if specDir != "" {
		projection, err = NewProjectionFromFolder(specDir)
		if err != nil {
			return nil, errors.Wrap(err, "Error parsing spec config folder")
		}
	} else {
		projection, err = NewProjectionFromFile(specFile)
		if err != nil {
			return nil, errors.Wrap(err, "Error parsing spec config file")
		}
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
	blockCol := make(map[string]types.SQLTableColumn)
	txCol := make(map[string]types.SQLTableColumn)

	// block table
	blockCol[types.BlockHeightLabel] = types.SQLTableColumn{
		Name:    types.SQLColumnLabelHeight,
		Type:    types.SQLColumnTypeVarchar,
		Length:  100,
		Primary: true,
		Order:   1,
	}

	blockCol[types.BlockHeaderLabel] = types.SQLTableColumn{
		Name:    types.SQLColumnLabelBlockHeader,
		Type:    types.SQLColumnTypeJSON,
		Primary: false,
		Order:   2,
	}

	// transaction table
	txCol[types.BlockHeightLabel] = types.SQLTableColumn{
		Name:    types.SQLColumnLabelHeight,
		Type:    types.SQLColumnTypeVarchar,
		Length:  100,
		Primary: true,
		Order:   1,
	}

	txCol[types.TxTxHashLabel] = types.SQLTableColumn{
		Name:    types.SQLColumnLabelTxHash,
		Type:    types.SQLColumnTypeVarchar,
		Length:  txs.HashLengthHex,
		Primary: true,
		Order:   2,
	}

	txCol[types.TxIndexLabel] = types.SQLTableColumn{
		Name:    types.SQLColumnLabelIndex,
		Type:    types.SQLColumnTypeNumeric,
		Length:  0,
		Primary: false,
		Order:   3,
	}

	txCol[types.TxTxTypeLabel] = types.SQLTableColumn{
		Name:    types.SQLColumnLabelTxType,
		Type:    types.SQLColumnTypeVarchar,
		Length:  100,
		Primary: false,
		Order:   4,
	}

	txCol[types.TxEnvelopeLabel] = types.SQLTableColumn{
		Name:    types.SQLColumnLabelEnvelope,
		Type:    types.SQLColumnTypeJSON,
		Primary: false,
		Order:   5,
	}

	txCol[types.TxEventsLabel] = types.SQLTableColumn{
		Name:    types.SQLColumnLabelEvents,
		Type:    types.SQLColumnTypeJSON,
		Primary: false,
		Order:   6,
	}

	txCol[types.TxResultLabel] = types.SQLTableColumn{
		Name:    types.SQLColumnLabelResult,
		Type:    types.SQLColumnTypeJSON,
		Primary: false,
		Order:   7,
	}

	txCol[types.TxReceiptLabel] = types.SQLTableColumn{
		Name:    types.SQLColumnLabelReceipt,
		Type:    types.SQLColumnTypeJSON,
		Primary: false,
		Order:   8,
	}

	txCol[types.TxExceptionLabel] = types.SQLTableColumn{
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
