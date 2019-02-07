package sqlsol_test

import (
	"strings"
	"testing"

	"github.com/hyperledger/burrow/vent/sqlsol"
	"github.com/hyperledger/burrow/vent/test"
	"github.com/hyperledger/burrow/vent/types"
	"github.com/stretchr/testify/require"
)

func TestNewProjection(t *testing.T) {
	t.Run("returns an error if the json is malformed", func(t *testing.T) {
		badJSON := test.BadJSONConfFile(t)

		byteValue := []byte(badJSON)
		_, err := sqlsol.NewProjectionFromBytes(byteValue)
		require.Error(t, err)
	})

	t.Run("returns an error if needed json fields are missing", func(t *testing.T) {
		missingFields := test.MissingFieldsJSONConfFile(t)

		byteValue := []byte(missingFields)
		_, err := sqlsol.NewProjectionFromBytes(byteValue)
		require.Error(t, err)
	})

	t.Run("successfully builds table structure based on json events config", func(t *testing.T) {
		goodJSON := test.GoodJSONConfFile(t)

		byteValue := []byte(goodJSON)
		tableStruct, err := sqlsol.NewProjectionFromBytes(byteValue)
		require.NoError(t, err)

		// columns map
		col, err := tableStruct.GetColumn("UserAccounts", "userName")
		require.NoError(t, err)
		require.Equal(t, false, col.Primary)
		require.Equal(t, types.SQLColumnTypeText, col.Type)
		require.Equal(t, "username", col.Name)

		col, err = tableStruct.GetColumn("UserAccounts", "userAddress")
		require.NoError(t, err)
		require.Equal(t, true, col.Primary)
		require.Equal(t, types.SQLColumnTypeVarchar, col.Type)
		require.Equal(t, "address", col.Name)

		col, err = tableStruct.GetColumn("UserAccounts", "txHash")
		require.NoError(t, err)
		require.Equal(t, false, col.Primary)
		require.Equal(t, types.SQLColumnTypeVarchar, col.Type)
		require.Equal(t, "_txhash", col.Name)
		require.Equal(t, 2, col.Order)

		col, err = tableStruct.GetColumn("UserAccounts", "eventName")
		require.NoError(t, err)
		require.Equal(t, false, col.Primary)
		require.Equal(t, types.SQLColumnTypeVarchar, col.Type)
		require.Equal(t, "_eventname", col.Name)
		require.Equal(t, 4, col.Order)
	})

	t.Run("returns an error if the event type of a given column is unknown", func(t *testing.T) {
		typeUnknownJSON := test.UnknownTypeJSONConfFile(t)

		byteValue := []byte(typeUnknownJSON)
		_, err := sqlsol.NewProjectionFromBytes(byteValue)
		require.Error(t, err)
	})

	t.Run("returns an error if there are duplicated column names for a given table in json file", func(t *testing.T) {
		duplicatedColNameJSON := test.DuplicatedColNameJSONConfFile(t)

		byteValue := []byte(duplicatedColNameJSON)
		_, err := sqlsol.NewProjectionFromBytes(byteValue)
		require.Error(t, err)
	})

}

func TestGetColumn(t *testing.T) {
	goodJSON := test.GoodJSONConfFile(t)

	byteValue := []byte(goodJSON)
	tableStruct, _ := sqlsol.NewProjectionFromBytes(byteValue)

	t.Run("successfully gets the mapping column info for a given table & column name", func(t *testing.T) {
		column, err := tableStruct.GetColumn("TEST_TABLE", "blocknum")
		require.NoError(t, err)
		require.Equal(t, strings.ToLower("Block"), column.Name)
		require.Equal(t, types.SQLColumnTypeBigInt, column.Type)
		require.Equal(t, false, column.Primary)

		column, err = tableStruct.GetColumn("TEST_TABLE", "instance")
		require.NoError(t, err)
		require.Equal(t, strings.ToLower("Instance"), column.Name)
		require.Equal(t, types.SQLColumnTypeBigInt, column.Type)
		require.Equal(t, false, column.Primary)

	})

	t.Run("unsuccessfully gets the mapping column info for a non existent table name", func(t *testing.T) {
		_, err := tableStruct.GetColumn("NOT_EXISTS", "userName")
		require.Error(t, err)
	})

	t.Run("unsuccessfully gets the mapping column info for a non existent column name", func(t *testing.T) {
		_, err := tableStruct.GetColumn("UpdateUserAccount", "NOT_EXISTS")
		require.Error(t, err)
	})
}

func TestGetTables(t *testing.T) {
	goodJSON := test.GoodJSONConfFile(t)

	byteValue := []byte(goodJSON)
	tableStruct, _ := sqlsol.NewProjectionFromBytes(byteValue)

	t.Run("successfully returns event tables structures", func(t *testing.T) {
		tables := tableStruct.GetTables()
		require.Equal(t, 2, len(tables))
		require.Equal(t, "useraccounts", tables["UserAccounts"].Name)
		require.Equal(t, "LOG0 = 'UserAccounts'", tables["UserAccounts"].Filter)

	})
}

func TestGetEventSpec(t *testing.T) {
	goodJSON := test.GoodJSONConfFile(t)

	byteValue := []byte(goodJSON)
	tableStruct, _ := sqlsol.NewProjectionFromBytes(byteValue)

	t.Run("successfully returns event specification structures", func(t *testing.T) {
		eventSpec := tableStruct.GetEventSpec()
		require.Equal(t, 2, len(eventSpec))
		require.Equal(t, "LOG0 = 'UserAccounts'", eventSpec[0].Filter)
		require.Equal(t, "UserAccounts", eventSpec[0].TableName)

		require.Equal(t, "Log1Text = 'EVENT_TEST'", eventSpec[1].Filter)
		require.Equal(t, "TEST_TABLE", eventSpec[1].TableName)
	})
}
