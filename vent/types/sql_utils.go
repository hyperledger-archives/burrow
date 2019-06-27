package types

import "github.com/hyperledger/burrow/logging"

// SQLConnection stores parameters to build a new db connection & initialize the database
type SQLConnection struct {
	DBAdapter string
	DBURL     string
	DBSchema  string
	Log       *logging.Logger
}

// SQLCleanDBQuery stores queries needed to clean the database
type SQLCleanDBQuery struct {
	SelectChainIDQry    string
	DeleteChainIDQry    string
	InsertChainIDQry    string
	SelectDictionaryQry string
	DeleteDictionaryQry string
	DeleteLogQry        string
}
