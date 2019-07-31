package sqlsol

import (
	"fmt"

	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/vent/types"
)

type SpecOpt uint64

const (
	Block SpecOpt = 1 << iota
	Tx
)

const (
	None    SpecOpt = 0
	BlockTx         = Block | Tx
)

// SpecLoader loads spec files and parses them
func SpecLoader(specFileOrDirs []string, opts SpecOpt) (*Projection, error) {
	var projection *Projection
	var err error

	if len(specFileOrDirs) == 0 {
		return nil, fmt.Errorf("please provide a spec file or directory")
	}

	projection, err = NewProjectionFromFolder(specFileOrDirs...)
	if err != nil {
		return nil, fmt.Errorf("error parsing spec: %v", err)
	}

	// add block & tx to tables definition
	if Block&opts > 0 {
		for k, v := range blockTables() {
			projection.Tables[k] = v
		}
	}
	if Tx&opts > 0 {
		for k, v := range txTables() {
			projection.Tables[k] = v
		}
	}

	return projection, nil
}

// getBlockTxTablesDefinition returns block & transaction structures
func blockTables() types.EventTables {
	return types.EventTables{
		tables.Block: &types.SQLTable{
			Name: tables.Block,
			Columns: []*types.SQLTableColumn{
				{
					Name:    columns.Height,
					Type:    types.SQLColumnTypeVarchar,
					Length:  100,
					Primary: true,
				},
				{
					Name:    columns.BlockHeader,
					Type:    types.SQLColumnTypeJSON,
					Primary: false,
				},
			},
		},
	}
}

func txTables() types.EventTables {
	return types.EventTables{
		tables.Tx: &types.SQLTable{
			Name: tables.Tx,
			Columns: []*types.SQLTableColumn{
				// transaction table
				{
					Name:    columns.Height,
					Type:    types.SQLColumnTypeVarchar,
					Length:  100,
					Primary: true,
				},
				{
					Name:    columns.TxHash,
					Type:    types.SQLColumnTypeVarchar,
					Length:  txs.HashLengthHex,
					Primary: true,
				},
				{
					Name:    columns.TxIndex,
					Type:    types.SQLColumnTypeNumeric,
					Length:  0,
					Primary: false,
				},
				{
					Name:    columns.TxType,
					Type:    types.SQLColumnTypeVarchar,
					Length:  100,
					Primary: false,
				},
				{
					Name:    columns.Envelope,
					Type:    types.SQLColumnTypeJSON,
					Primary: false,
				},
				{
					Name:    columns.Events,
					Type:    types.SQLColumnTypeJSON,
					Primary: false,
				},
				{
					Name:    columns.Result,
					Type:    types.SQLColumnTypeJSON,
					Primary: false,
				},
				{
					Name:    columns.Receipt,
					Type:    types.SQLColumnTypeJSON,
					Primary: false,
				},
				{
					Name:    columns.Exception,
					Type:    types.SQLColumnTypeJSON,
					Primary: false,
				},
			},
		},
	}
}
