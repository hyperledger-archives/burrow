package config

import (
	"time"

	"github.com/hyperledger/burrow/vent/sqlsol"

	"github.com/hyperledger/burrow/vent/types"
)

const DefaultPostgresDBURL = "postgres://postgres@localhost:5432/postgres?sslmode=disable"

// VentConfig is a set of configuration parameters
type VentConfig struct {
	DBAdapter      string
	DBURL          string
	DBSchema       string
	GRPCAddr       string
	HTTPAddr       string
	LogLevel       string
	SpecFileOrDirs []string
	AbiFileOrDirs  []string
	SpecOpt        sqlsol.SpecOpt
	// Announce status every AnnouncePeriod
	AnnounceEvery time.Duration
}

// DefaultFlags returns a configuration with default values
func DefaultVentConfig() *VentConfig {
	return &VentConfig{
		DBAdapter:     types.PostgresDB,
		DBURL:         DefaultPostgresDBURL,
		DBSchema:      "vent",
		GRPCAddr:      "localhost:10997",
		HTTPAddr:      "0.0.0.0:8080",
		LogLevel:      "debug",
		SpecOpt:       sqlsol.None,
		AnnounceEvery: time.Second * 5,
	}
}
