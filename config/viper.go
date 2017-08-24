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

package config

import (
	"fmt"

	"bytes"

	"github.com/spf13/viper"
)

// Safely get the subtree from a viper config, returning an error if it could not
// be obtained for any reason.
func ViperSubConfig(conf *viper.Viper, configSubtreePath string) (subConfig *viper.Viper, err error) {
	// Viper internally panics if `moduleName` contains an disallowed
	// character (eg, a dash).
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Viper panicked trying to read config subtree: %s",
				configSubtreePath)
		}
	}()
	if !conf.IsSet(configSubtreePath) {
		return nil, fmt.Errorf("Failed to read config subtree: %s",
			configSubtreePath)
	}
	subConfig = conf.Sub(configSubtreePath)
	if subConfig == nil {
		return nil, fmt.Errorf("Failed to read config subtree: %s",
			configSubtreePath)
	}
	return subConfig, err
}

// Read in TOML Viper config from bytes
func ReadViperConfig(configBytes []byte) (*viper.Viper, error) {
	buf := bytes.NewBuffer(configBytes)
	conf := viper.New()
	viper.SetConfigType("toml")
	err := conf.ReadConfig(buf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
