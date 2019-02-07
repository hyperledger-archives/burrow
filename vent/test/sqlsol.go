package test

import (
	"testing"
)

// GoodJSONConfFile sets a good json file to be used in projection tests
func GoodJSONConfFile(t *testing.T) string {
	t.Helper()

	goodJSONConfFile := `[
		{
			"TableName" : "UserAccounts",
			"Filter" : "LOG0 = 'UserAccounts'",
			"DeleteMarkerField": "__DELETE__",
			"Fields"  : {
				"userAddress" : {"ColumnName" : "address", "Type": "address", "Primary" : true},
				"userName": {"ColumnName" : "username", "Type": "string", "Primary" : false},
				"userId": {"ColumnName" : "userid", "Type": "uint256", "Primary" : false},
				"userBool": {"ColumnName" : "userbool", "Type": "bool", "Primary" : false}
			}
		},
		{
		"TableName" : "TEST_TABLE",
		"Filter" : "Log1Text = 'EVENT_TEST'",
		"DeleteMarkerField": "__DELETE__",
		"Fields"  : {
			"key"		: {"ColumnName" : "Index",    "Type": "uint256", "Primary" : true},
			"blocknum"  : {"ColumnName" : "Block",    "Type": "uint256", "Primary" : false},
			"somestr"	: {"ColumnName" : "String",   "Type": "string", "Primary" : false},
			"instance" 	: {"ColumnName" : "Instance", "Type": "uint", "Primary" : false}
		}
	}
	]`

	return goodJSONConfFile
}

// MissingFieldsJSONConfFile sets a json file with missing fields to be used in projection tests
func MissingFieldsJSONConfFile(t *testing.T) string {
	t.Helper()

	missingFieldsJSONConfFile := `[
		{
			"TableName" : "UserAccounts",
			"Fields"  : {
				"userAddress" : {"ColumnName" : "address", "Primary" : true},
				"userName": {"ColumnName" : "username", "Primary" : false}
			}
		}
	]`

	return missingFieldsJSONConfFile
}

// UnknownTypeJSONConfFile sets a json file with unknown column types to be used in projection tests
func UnknownTypeJSONConfFile(t *testing.T) string {
	t.Helper()

	unknownTypeJSONConfFile := `[
		{
			"TableName" : "UserAccounts",
			"Filter" : "LOG0 = 'UserAccounts'",
			"DeleteMarkerField": "__DELETE__",
			"Event"  : {
				"anonymous": false,
				"inputs": [{
					"indexed": false,
					"ColumnName": "userName",
					"Type": "typeunknown"
				}, {
					"indexed": false,
					"ColumnName": "userAddress",
					"Type": "address"
				}, {
					"indexed": false,
					"ColumnName": "UnimportantInfo",
					"Type": "uint"
				}],
				"ColumnName": "UpdateUserAccount",
				"Type": "event"
			},
			"Fields"  : {
				"userAddress" : {"ColumnName" : "address", "Primary" : true},
				"userName": {"ColumnName" : "username", "Primary" : false}
			}
		},
		{
			"TableName" : "EventTest",
			"Filter" : "LOG0 = 'EventTest'",
			"DeleteMarkerField": "__DELETE__",
			"Event"  : {
				"anonymous": false,
				"inputs": [{
					"indexed": false,
					"ColumnName": "ColumnName",
					"Type": "typeunknown"
				}, {
					"indexed": false,
					"ColumnName": "description",
					"Type": "string"
				}, {
					"indexed": false,
					"ColumnName": "UnimportantInfo",
					"Type": "uint"
				}],
				"ColumnName": "TEST_EVENTS",
				"Type": "event"
			},
			"Fields"  : {
				"ColumnName" : {"ColumnName" : "testname", "Primary" : true},
				"description": {"ColumnName" : "testdescription", "Primary" : false}
			}
		}
	]`

	return unknownTypeJSONConfFile
}

// BadJSONConfFile sets a malformed json file to be used in projection tests
func BadJSONConfFile(t *testing.T) string {
	t.Helper()

	badJSONConfFile := `[
		{
			"TableName" : "UserAccounts",
			"Event"  : {
				"anonymous": false,
				"inputs": [{
					"indexed": false,
					"ColumnName": "userName",
					"Type": "string"
				}, {
					"indexed": false,
					"ColumnName": "userAddress",
					"Type": "address"
				}, {
					"indexed": false,
					"ColumnName": "UnimportantInfo",
					"Type": "uint"
				}],
				"ColumnName": "UpdateUserAccount",
			},
			"Fields"  : {
				"userAddress" : {"ColumnName" : "address", "Primary" : true},
				"userName": {"ColumnName" : "username", "Primary" : false}
	]`

	return badJSONConfFile
}

// DuplicatedColNameJSONConfFile sets a good json file but with duplicated column names for a given table
func DuplicatedColNameJSONConfFile(t *testing.T) string {
	t.Helper()

	duplicatedColNameJSONConfFile := `[
		{
			"TableName" : "DUPLICATED_COLUMN",
			"Filter" : "LOG0 = 'UserAccounts'",
			"DeleteMarkerField": "__DELETE__",
			"Fields"  : {
				"userAddress" : {"ColumnName" : "address", "Primary" : true},
				"userName": {"ColumnName" : "duplicated", "Primary" : false},
				"userId": {"ColumnName" : "userid", "Primary" : false},
				"userBool": {"ColumnName" : "duplicated", "Primary" : false}
			}
	}
	]`

	return duplicatedColNameJSONConfFile
}
