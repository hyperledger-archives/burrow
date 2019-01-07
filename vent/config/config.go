package config

import (
	"github.com/hyperledger/burrow/vent/types"
)

// Flags is a set of configuration parameters
type Flags struct {
	DBAdapter string
	DBURL     string
	DBSchema  string
	GRPCAddr  string
	HTTPAddr  string
	LogLevel  string
	SpecFile  string
	SpecDir   string
	AbiFile   string
	AbiDir    string
	DBBlockTx bool
}

// DefaultFlags returns a configuration with default values
func DefaultFlags() *Flags {
	return &Flags{
		DBAdapter: types.PostgresDB,
		DBURL:     "postgres://user:pass@localhost:5432/vent?sslmode=disable",
		DBSchema:  "vent",
		GRPCAddr:  "localhost:10997",
		HTTPAddr:  "0.0.0.0:8080",
		LogLevel:  "debug",
		SpecFile:  "",
		SpecDir:   "",
		AbiFile:   "",
		AbiDir:    "",
		DBBlockTx: false,
	}
}
