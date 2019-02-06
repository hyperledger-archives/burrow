package types

import "github.com/hyperledger/burrow/vent/logger"

// SQLConnection stores parameters to build a new db connection & initialize the database
type SQLConnection struct {
	DBAdapter     string
	DBURL         string
	DBSchema      string
	Log           *logger.Logger
	ChainID       string
	BurrowVersion string
}

// SQLCleanDBQuery stores queries needed to clean the database
type SQLCleanDBQuery struct {
	SelectChainIDQry    string
	DeleteChainIDQry    string
	InsertChainIDQry    string
	SelectDictionaryQry string
	DeleteDictionaryQry string
	DeleteLogQry        string
	DeleteErrorsQry     string
}
