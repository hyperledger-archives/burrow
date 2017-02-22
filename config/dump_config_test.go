// +build dumpconfig

// Space above matters
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
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

// This is a little convenience for getting a config file dump. Just run:
// go test -tags dumpconfig ./config
// This pseudo test won't run unless the dumpconfig tag is
func TestDumpConfig(t *testing.T) {
	bs, err := GetExampleConfigFileBytes()
	assert.NoError(t, err, "Should be able to create example config")
	ioutil.WriteFile("config_dump.toml", bs, 0644)
}
