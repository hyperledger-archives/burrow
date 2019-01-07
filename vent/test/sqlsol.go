package test

import (
	"testing"
)

// GoodJSONConfFile sets a good json file to be used in parser tests
func GoodJSONConfFile(t *testing.T) string {
	t.Helper()

	goodJSONConfFile := `[
		{
			"TableName" : "UserAccounts",
			"Filter" : "LOG0 = 'UserAccounts'",
			"DeleteFilter": "CRUD_ACTION = 'delete'",
			"Columns"  : {
				"userAddress" : {"name" : "address", "type": "address", "primary" : true},
				"userName": {"name" : "username", "type": "string", "primary" : false},
				"userId": {"name" : "userid", "type": "uint256", "primary" : false},
				"userBool": {"name" : "userbool", "type": "bool", "primary" : false}
			}
		},
		{
		"TableName" : "TEST_TABLE",
		"Filter" : "Log1Text = 'EVENT_TEST'",
		"DeleteFilter": "CRUD_ACTION = 'delete'",
		"Columns"  : {
			"key"		: {"name" : "Index",    "type": "uint256", "primary" : true},
			"blocknum"  : {"name" : "Block",    "type": "uint256", "primary" : false},
			"somestr"	: {"name" : "String",   "type": "string", "primary" : false},
			"instance" 	: {"name" : "Instance", "type": "uint", "primary" : false}
		}
	}
	]`

	return goodJSONConfFile
}

// MissingFieldsJSONConfFile sets a json file with missing fields to be used in parser tests
func MissingFieldsJSONConfFile(t *testing.T) string {
	t.Helper()

	missingFieldsJSONConfFile := `[
		{
			"TableName" : "UserAccounts",
			"Columns"  : {
				"userAddress" : {"name" : "address", "primary" : true},
				"userName": {"name" : "username", "primary" : false}
			}
		}
	]`

	return missingFieldsJSONConfFile
}

// UnknownTypeJSONConfFile sets a json file with unknown column types to be used in parser tests
func UnknownTypeJSONConfFile(t *testing.T) string {
	t.Helper()

	unknownTypeJSONConfFile := `[
		{
			"TableName" : "UserAccounts",
			"Filter" : "LOG0 = 'UserAccounts'",
			"DeleteFilter": "CRUD_ACTION = 'delete'",
			"Event"  : {
				"anonymous": false,
				"inputs": [{
					"indexed": false,
					"name": "userName",
					"type": "typeunknown"
				}, {
					"indexed": false,
					"name": "userAddress",
					"type": "address"
				}, {
					"indexed": false,
					"name": "UnimportantInfo",
					"type": "uint"
				}],
				"name": "UpdateUserAccount",
				"type": "event"
			},
			"Columns"  : {
				"userAddress" : {"name" : "address", "primary" : true},
				"userName": {"name" : "username", "primary" : false}
			}
		},
		{
			"TableName" : "EventTest",
			"Filter" : "LOG0 = 'EventTest'",
			"DeleteFilter": "CRUD_ACTION = 'delete'",
			"Event"  : {
				"anonymous": false,
				"inputs": [{
					"indexed": false,
					"name": "name",
					"type": "typeunknown"
				}, {
					"indexed": false,
					"name": "description",
					"type": "string"
				}, {
					"indexed": false,
					"name": "UnimportantInfo",
					"type": "uint"
				}],
				"name": "TEST_EVENTS",
				"type": "event"
			},
			"Columns"  : {
				"name" : {"name" : "testname", "primary" : true},
				"description": {"name" : "testdescription", "primary" : false}
			}
		}
	]`

	return unknownTypeJSONConfFile
}

// BadJSONConfFile sets a malformed json file to be used in parser tests
func BadJSONConfFile(t *testing.T) string {
	t.Helper()

	badJSONConfFile := `[
		{
			"TableName" : "UserAccounts",
			"Event"  : {
				"anonymous": false,
				"inputs": [{
					"indexed": false,
					"name": "userName",
					"type": "string"
				}, {
					"indexed": false,
					"name": "userAddress",
					"type": "address"
				}, {
					"indexed": false,
					"name": "UnimportantInfo",
					"type": "uint"
				}],
				"name": "UpdateUserAccount",
			},
			"Columns"  : {
				"userAddress" : {"name" : "address", "primary" : true},
				"userName": {"name" : "username", "primary" : false}
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
			"DeleteFilter": "CRUD_ACTION = 'delete'",
			"Columns"  : {
				"userAddress" : {"name" : "address", "primary" : true},
				"userName": {"name" : "duplicated", "primary" : false},
				"userId": {"name" : "userid", "primary" : false},
				"userBool": {"name" : "duplicated", "primary" : false}
			}
	}
	]`

	return duplicatedColNameJSONConfFile
}
