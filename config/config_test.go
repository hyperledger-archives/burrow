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
	"bytes"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// Since the logic for generating configuration files (in eris-cm) is split from
// the logic for consuming them
func TestGeneratedConfigIsUsable(t *testing.T) {
	bs, err := GetExampleConfigFileBytes()
	assert.NoError(t, err, "Should be able to create example config")
	buf := bytes.NewBuffer(bs)
	conf := viper.New()
	viper.SetConfigType("toml")
	err = conf.ReadConfig(buf)
	assert.NoError(t, err, "Should be able to read example config into Viper")
}
