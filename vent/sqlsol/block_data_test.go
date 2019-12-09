package sqlsol_test

import (
	"testing"

	"github.com/hyperledger/burrow/vent/sqlsol"
	"github.com/hyperledger/burrow/vent/types"
	"github.com/stretchr/testify/require"
)

func TestSetRow(t *testing.T) {
	t.Run("successfully sets a new data row", func(t *testing.T) {
		values := make(map[string]interface{})
		values["c1"] = "v1"
		values["c2"] = "v2"

		blockData := sqlsol.NewBlockData(44)
		blockData.AddRow("TEST_TABLE", types.EventDataRow{Action: types.ActionUpsert, RowData: values})

		rows, err := blockData.GetRows("TEST_TABLE")
		require.NoError(t, err)
		require.Equal(t, 1, len(rows))
		require.Equal(t, "v1", rows[0].RowData["c1"])
		require.Equal(t, "v2", rows[0].RowData["c2"])
	})
}

func TestGetBlockData(t *testing.T) {
	t.Run("successfully gets block data", func(t *testing.T) {
		blockData := sqlsol.NewBlockData(2)
		blk := blockData.Data
		require.EqualValues(t, 2, blk.BlockHeight)
	})
}
