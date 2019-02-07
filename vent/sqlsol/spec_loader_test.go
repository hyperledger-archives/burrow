package sqlsol_test

import (
	"os"
	"testing"

	"github.com/hyperledger/burrow/vent/sqlsol"
	"github.com/hyperledger/burrow/vent/types"
	"github.com/stretchr/testify/require"
)

func TestSpecLoader(t *testing.T) {
	specFile := os.Getenv("GOPATH") + "/src/github.com/hyperledger/burrow/vent/test/sqlsol_example.json"
	dBBlockTx := true
	t.Run("successfully add block and transaction tables to event structures", func(t *testing.T) {
		projection, err := sqlsol.SpecLoader(specFile, dBBlockTx)
		require.NoError(t, err)

		require.Equal(t, 4, len(projection.Tables))
		require.Equal(t, types.SQLBlockTableName, projection.Tables[types.SQLBlockTableName].Name)
		require.Equal(t, "_height", projection.Tables[types.SQLBlockTableName].Columns["height"].Name)
		require.Equal(t, types.SQLTxTableName, projection.Tables[types.SQLTxTableName].Name)
		require.Equal(t, "_txhash", projection.Tables[types.SQLTxTableName].Columns["txHash"].Name)
	})
}
