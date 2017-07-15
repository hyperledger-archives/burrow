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

// config defines simple types in a separate package to avoid cyclical imports

import (
	viper "github.com/spf13/viper"
)

type ModuleConfig struct {
	Module      string
	Name        string
	WorkDir     string
	DataDir     string
	RootDir     string
	ChainId     string
	GenesisFile string
	Config      *viper.Viper
}
