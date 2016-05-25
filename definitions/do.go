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
  viper "github.com/spf13/viper"
)

type Do struct {
  // Persistent flags not reflected in the configuration files
  // only set through command line flags or environment variables
	Debug        bool     // ERIS_DB_DEBUG
	Verbose      bool     // ERIS_DB_VERBOSE
  WorkDir      string
  // Capital configuration options explicitly extracted from the Viper config
	ChainId      string   // has to be set to non-empty string,
                        // uniquely identifying the chain.
	// ChainType    string
	// CSV          string
	// AccountTypes []string
	// Zip          bool
	// Tarball      bool
	Config       *viper.Viper
	// Accounts     []*Account
	// Result       string
}

func NowDo() *Do {
	do := new(Do)
	do.Debug = false
	do.Verbose = false
	// the default value for output is set to true in cmd/eris-db.go;
	// avoid double setting it here though
  do.WorkDir = ""
  do.ChainId = ""
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
