// Copyright 2015, 2016 Eris Industries (UK) Ltd.
// This file is part of Eris-RT

// Eris-RT is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Eris-RT is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Eris-RT.  If not, see <http://www.gnu.org/licenses/>.

package definitions

import (
	"os"
	"path"

	viper "github.com/spf13/viper"

	util "github.com/eris-ltd/eris-db/util"
)

type Do struct {
	// Persistent flags not reflected in the configuration files
	// only set through command line flags or environment variables
	Debug   bool // ERIS_DB_DEBUG
	Verbose bool // ERIS_DB_VERBOSE

	// Work directory is the root directory for Eris-DB to act in
	WorkDir string // ERIS_DB_WORKDIR
	// Data directory is defaulted to WorkDir + `/data`.
	// If Eris-CLI maps a data container, DataDir is intended to point
	// to that mapped data directory.
	DataDir string // ERIS_DB_DATADIR

	// Capital configuration options explicitly extracted from the Viper config
	ChainId string // has to be set to non-empty string,
	// uniquely identifying the chain.
	GenesisFile string
	// ChainType    string
	// CSV          string
	// AccountTypes []string
	// Zip          bool
	// Tarball      bool
	DisableRpc bool
	Config     *viper.Viper
	// Accounts     []*Account
	// Result       string
}

func NewDo() *Do {
	do := new(Do)
	do.Debug = false
	do.Verbose = false
	do.WorkDir = ""
	do.DataDir = ""
	do.ChainId = ""
	do.GenesisFile = ""
	do.DisableRpc = false
	do.Config = viper.New()
	return do
}

// ReadConfig uses Viper to set the configuration file name, file format
// where Eris-DB currently only uses `toml`.
// The search directory is explicitly limited to a single location to
// minimise the chance of loading the wrong configuration file.
func (d *Do) ReadConfig(directory string, name string, configType string) error {
	// name of the configuration file without extension
	d.Config.SetConfigName(name)
	// Eris-DB currently only uses "toml"
	d.Config.SetConfigType(configType)
	// look for configuration file in the working directory
	d.Config.AddConfigPath(directory)
	return d.Config.ReadInConfig()
}

// InitialiseDataDirectory will default to WorkDir/data if DataDir is empty
func (d *Do) InitialiseDataDirectory() error {
	if d.DataDir == "" {
		d.DataDir = path.Join(d.WorkDir, "data")
	}
	return util.EnsureDir(d.DataDir, os.ModePerm)
}
