package sqlsol_test

import (
	"os"
	"testing"

	"github.com/hyperledger/burrow/vent/sqlsol"
	"github.com/hyperledger/burrow/vent/types"
	"github.com/stretchr/testify/require"
)

var tables = types.DefaultSQLTableNames

func TestSpecLoader(t *testing.T) {
	specFile := []string{os.Getenv("GOPATH") + "/src/github.com/hyperledger/burrow/vent/test/sqlsol_view.json"}
	t.Run("successfully add block and transaction tables to event structures", func(t *testing.T) {
		projection, err := sqlsol.SpecLoader(specFile, sqlsol.BlockTx)
		require.NoError(t, err)

		require.Equal(t, 4, len(projection.Tables))

		require.Equal(t, tables.Block, projection.Tables[tables.Block].Name)

		require.Equal(t, columns.Height,
			projection.Tables[tables.Block].GetColumn(columns.Height).Name)

		require.Equal(t, tables.Tx, projection.Tables[tables.Tx].Name)

		require.Equal(t, columns.TxHash,
			projection.Tables[tables.Tx].GetColumn(columns.TxHash).Name)
	})
}
