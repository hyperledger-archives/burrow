// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package definitions

import (
	"os"
	"path"

	viper "github.com/spf13/viper"

	util "github.com/monax/eris-db/util"
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
