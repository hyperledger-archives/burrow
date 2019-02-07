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
			"DeleteFilter": "CRUD_ACTION = 'delete'",
			"Columns"  : {
				"userAddress" : {"Name" : "address", "Type": "address", "Primary" : true},
				"userName": {"Name" : "username", "Type": "string", "Primary" : false},
				"userId": {"Name" : "userid", "Type": "uint256", "Primary" : false},
				"userBool": {"Name" : "userbool", "Type": "bool", "Primary" : false}
			}
		},
		{
		"TableName" : "TEST_TABLE",
		"Filter" : "Log1Text = 'EVENT_TEST'",
		"DeleteFilter": "CRUD_ACTION = 'delete'",
		"Columns"  : {
			"key"		: {"Name" : "Index",    "Type": "uint256", "Primary" : true},
			"blocknum"  : {"Name" : "Block",    "Type": "uint256", "Primary" : false},
			"somestr"	: {"Name" : "String",   "Type": "string", "Primary" : false},
			"instance" 	: {"Name" : "Instance", "Type": "uint", "Primary" : false}
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
			"Columns"  : {
				"userAddress" : {"Name" : "address", "Primary" : true},
				"userName": {"Name" : "username", "Primary" : false}
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
			"DeleteFilter": "CRUD_ACTION = 'delete'",
			"Event"  : {
				"anonymous": false,
				"inputs": [{
					"indexed": false,
					"Name": "userName",
					"Type": "typeunknown"
				}, {
					"indexed": false,
					"Name": "userAddress",
					"Type": "address"
				}, {
					"indexed": false,
					"Name": "UnimportantInfo",
					"Type": "uint"
				}],
				"Name": "UpdateUserAccount",
				"Type": "event"
			},
			"Columns"  : {
				"userAddress" : {"Name" : "address", "Primary" : true},
				"userName": {"Name" : "username", "Primary" : false}
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
					"Name": "Name",
					"Type": "typeunknown"
				}, {
					"indexed": false,
					"Name": "description",
					"Type": "string"
				}, {
					"indexed": false,
					"Name": "UnimportantInfo",
					"Type": "uint"
				}],
				"Name": "TEST_EVENTS",
				"Type": "event"
			},
			"Columns"  : {
				"Name" : {"Name" : "testname", "Primary" : true},
				"description": {"Name" : "testdescription", "Primary" : false}
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
					"Name": "userName",
					"Type": "string"
				}, {
					"indexed": false,
					"Name": "userAddress",
					"Type": "address"
				}, {
					"indexed": false,
					"Name": "UnimportantInfo",
					"Type": "uint"
				}],
				"Name": "UpdateUserAccount",
			},
			"Columns"  : {
				"userAddress" : {"Name" : "address", "Primary" : true},
				"userName": {"Name" : "username", "Primary" : false}
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
				"userAddress" : {"Name" : "address", "Primary" : true},
				"userName": {"Name" : "duplicated", "Primary" : false},
				"userId": {"Name" : "userid", "Primary" : false},
				"userBool": {"Name" : "duplicated", "Primary" : false}
			}
	}
	]`

	return duplicatedColNameJSONConfFile
}
