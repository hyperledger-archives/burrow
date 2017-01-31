// Copyright 2015, 2016 Monax Industries (UK) Ltd.
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
