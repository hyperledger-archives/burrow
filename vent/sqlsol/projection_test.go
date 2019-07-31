package sqlsol_test

import (
	"testing"

	"github.com/hyperledger/burrow/vent/sqlsol"
	"github.com/hyperledger/burrow/vent/test"
	"github.com/hyperledger/burrow/vent/types"
	"github.com/stretchr/testify/require"
)

var columns = types.DefaultSQLColumnNames

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
		tableStruct, err := sqlsol.NewProjectionFromBytes([]byte(test.GoodJSONConfFile(t)))
		require.NoError(t, err)

		// columns map
		tableName := "UserAccounts"
		col, err := tableStruct.GetColumn(tableName, "username")
		require.NoError(t, err)
		require.Equal(t, false, col.Primary)
		require.Equal(t, types.SQLColumnTypeText, col.Type)
		require.Equal(t, "username", col.Name)

		col, err = tableStruct.GetColumn(tableName, "address")
		require.NoError(t, err)
		require.Equal(t, true, col.Primary)
		require.Equal(t, types.SQLColumnTypeVarchar, col.Type)
		require.Equal(t, "address", col.Name)

		col, err = tableStruct.GetColumn(tableName, columns.TxHash)
		require.NoError(t, err)
		require.Equal(t, false, col.Primary)
		require.Equal(t, types.SQLColumnTypeText, col.Type)

		col, err = tableStruct.GetColumn(tableName, columns.EventName)
		require.NoError(t, err)
		require.Equal(t, false, col.Primary)
		require.Equal(t, types.SQLColumnTypeText, col.Type)
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
	projection, err := sqlsol.NewProjectionFromBytes([]byte(test.GoodJSONConfFile(t)))
	require.NoError(t, err)

	t.Run("successfully gets the mapping column info for a given table & column name", func(t *testing.T) {
		column, err := projection.GetColumn("TEST_TABLE", "Block")
		require.NoError(t, err)
		require.Equal(t, "Block", column.Name)
		require.Equal(t, types.SQLColumnTypeBigInt, column.Type)
		require.Equal(t, false, column.Primary)

		column, err = projection.GetColumn("TEST_TABLE", "Instance")
		require.NoError(t, err)
		require.Equal(t, "Instance", column.Name)
		require.Equal(t, types.SQLColumnTypeBigInt, column.Type)
		require.Equal(t, false, column.Primary)

	})

	t.Run("unsuccessfully gets the mapping column info for a non existent table name", func(t *testing.T) {
		_, err := projection.GetColumn("NOT_EXISTS", "userName")
		require.Error(t, err)
	})

	t.Run("unsuccessfully gets the mapping column info for a non existent column name", func(t *testing.T) {
		_, err := projection.GetColumn("UpdateUserAccount", "NOT_EXISTS")
		require.Error(t, err)
	})
}

func TestGetTables(t *testing.T) {
	goodJSON := test.GoodJSONConfFile(t)

	byteValue := []byte(goodJSON)
	tableStruct, _ := sqlsol.NewProjectionFromBytes(byteValue)

	t.Run("successfully returns event tables structures", func(t *testing.T) {
		tables := tableStruct.Tables
		require.Equal(t, 2, len(tables))
		require.Equal(t, "UserAccounts", tables["UserAccounts"].Name)

	})
}

func TestGetEventSpec(t *testing.T) {
	goodJSON := test.GoodJSONConfFile(t)

	byteValue := []byte(goodJSON)
	tableStruct, _ := sqlsol.NewProjectionFromBytes(byteValue)

	t.Run("successfully returns event specification structures", func(t *testing.T) {
		eventSpec := tableStruct.EventSpec
		require.Equal(t, 2, len(eventSpec))
		require.Equal(t, "LOG0 = 'UserAccounts'", eventSpec[0].Filter)
		require.Equal(t, "UserAccounts", eventSpec[0].TableName)

		require.Equal(t, "Log1Text = 'EVENT_TEST'", eventSpec[1].Filter)
		require.Equal(t, "TEST_TABLE", eventSpec[1].TableName)
	})
}

func TestNewProjectionFromEventSpec(t *testing.T) {
	tableName := "BurnNotices"
	eventSpec := types.EventSpec{
		{
			TableName: tableName,
			Filter:    "LOG1Text = 'CIA/burn'",
			FieldMappings: []*types.EventFieldMapping{
				{
					Field:      "codename",
					Type:       types.EventFieldTypeString,
					ColumnName: "name",
					Notify:     []string{"burn"},
					Primary:    true,
				},
				{
					Field:      "burn",
					Type:       types.EventFieldTypeBool,
					ColumnName: "burnt",
					Notify:     []string{"burn"},
				},
				{
					Field:      "dairy",
					Type:       types.EventFieldTypeString,
					ColumnName: "coffee_milk",
					Notify:     []string{"mrs_doyle"},
				},
				{
					Field:      "datetime",
					Type:       types.EventFieldTypeInt,
					ColumnName: "time_changed",
					Notify:     []string{"last_heard", "mrs_doyle"},
				},
			},
		},
		{
			TableName: tableName,
			Filter:    "LOG1Text = 'MI5/burn'",
			FieldMappings: []*types.EventFieldMapping{
				{
					Field:      "codename",
					Type:       types.EventFieldTypeString,
					ColumnName: "name",
					Notify:     []string{"burn"},
					Primary:    true,
				},
				{
					Field:      "unreliable",
					Type:       types.EventFieldTypeBool,
					ColumnName: "burnt",
					Notify:     []string{"burn"},
				},
				{
					Field:      "sugars",
					Type:       types.EventFieldTypeInt,
					ColumnName: "tea_sugars",
					Notify:     []string{"mrs_doyle"},
				},
				{
					Field:      "milk",
					Type:       types.EventFieldTypeBool,
					ColumnName: "tea_milk",
					Notify:     []string{"mrs_doyle"},
				},
				{
					Field:      "datetime",
					Type:       types.EventFieldTypeInt,
					ColumnName: "time_changed",
					Notify:     []string{"last_heard", "mrs_doyle"},
				},
			},
		},
	}
	projection, err := sqlsol.NewProjectionFromEventSpec(eventSpec)
	require.NoError(t, err, "burn and unreliable field mappings should unify to single column")

	require.Equal(t, []string{"burnt", "name"}, projection.Tables[tableName].NotifyChannels["burn"])

	// Notify sugars on the burn channel
	field := eventSpec[1].GetFieldMapping("sugars")
	field.Notify = append(field.Notify, "burn")

	projection, err = sqlsol.NewProjectionFromEventSpec(eventSpec)
	require.NoError(t, err)
	require.Equal(t, []string{"burnt", "name", "tea_sugars"}, projection.Tables[tableName].NotifyChannels["burn"])

	// Create a column conflict between burn and unreliable fields (both map to burnt so the SQL column def must be identical)
	field = eventSpec[1].GetFieldMapping("unreliable")
	field.Primary = !field.Primary
	_, err = sqlsol.NewProjectionFromEventSpec(eventSpec)
	require.Error(t, err)
}

func TestWithNoPrimaryKey(t *testing.T) {
	tableName := "BurnNotices"
	eventSpec := types.EventSpec{
		{
			TableName:         tableName,
			Filter:            "LOG1Text = 'CIA/burn'",
			DeleteMarkerField: "__DELETE__",
			FieldMappings: []*types.EventFieldMapping{
				{
					Field:      "codename",
					Type:       types.EventFieldTypeString,
					ColumnName: "name",
					Notify:     []string{"burn"},
				},
				{
					Field:      "burn",
					Type:       types.EventFieldTypeBool,
					ColumnName: "burnt",
					Notify:     []string{"burn"},
				},
				{
					Field:      "dairy",
					Type:       types.EventFieldTypeString,
					ColumnName: "coffee_milk",
					Notify:     []string{"mrs_doyle"},
				},
				{
					Field:      "datetime",
					Type:       types.EventFieldTypeInt,
					ColumnName: "time_changed",
					Notify:     []string{"last_heard", "mrs_doyle"},
				},
			},
		},
	}

	_, err := sqlsol.NewProjectionFromEventSpec(eventSpec)
	require.Error(t, err, "no DeleteMarkerField allowed if no primary key on")

	// Try again and now check that the right fields are primary
	eventSpec[0].DeleteMarkerField = ""

	projection, err := sqlsol.NewProjectionFromEventSpec(eventSpec)
	require.NoError(t, err, "projection with no primary key should be allowed")

	for _, c := range projection.Tables[tableName].Columns {
		switch c.Name {
		case columns.ChainID:
			require.Equal(t, true, c.Primary)
		case columns.Height:
			require.Equal(t, true, c.Primary)
		case columns.TxIndex:
			require.Equal(t, true, c.Primary)
		case columns.EventIndex:
			require.Equal(t, true, c.Primary)
		default:
			require.Equal(t, false, c.Primary)
		}
	}

	eventSpec = types.EventSpec{
		{
			TableName: tableName,
			Filter:    "LOG1Text = 'CIA/burn'",
			FieldMappings: []*types.EventFieldMapping{
				{
					Field:      "codename",
					Type:       types.EventFieldTypeString,
					ColumnName: "name",
					Notify:     []string{"burn"},
					Primary:    true,
				},
				{
					Field:      "burn",
					Type:       types.EventFieldTypeBool,
					ColumnName: "burnt",
					Notify:     []string{"burn"},
				},
				{
					Field:      "dairy",
					Type:       types.EventFieldTypeString,
					ColumnName: "coffee_milk",
					Notify:     []string{"mrs_doyle"},
				},
				{
					Field:      "datetime",
					Type:       types.EventFieldTypeInt,
					ColumnName: "time_changed",
					Notify:     []string{"last_heard", "mrs_doyle"},
				},
			},
		},
	}

	projection, err = sqlsol.NewProjectionFromEventSpec(eventSpec)
	require.NoError(t, err, "projection with primary key should be allowed")

	for _, c := range projection.Tables[tableName].Columns {
		require.Equal(t, c.Name == "name", c.Primary)
	}
}
