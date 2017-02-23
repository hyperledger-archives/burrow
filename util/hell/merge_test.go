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

package hell

import (
	"strings"
	"testing"

	"github.com/Masterminds/glide/cfg"
	"github.com/stretchr/testify/assert"
)

const baseLockYml = `imports:
- name: github.com/gogo/protobuf
  version: 82d16f734d6d871204a3feb1a73cb220cc92574c
- name: github.com/tendermint/tendermint
  version: aaea0c5d2e3ecfbf29f2608f9d43649ec7f07f50
  subpackages:
  - node
  - proxy
  - types
  - version
  - consensus
  - rpc/core/types
  - blockchain
  - mempool
  - rpc/core
  - state
`
const overrideLockYml = `imports:
- name: github.com/tendermint/tendermint
  version: 764091dfbb035f1b28da4b067526e04c6a849966
  subpackages:
  - benchmarks
  - proxy
  - types
  - version
`
const expectedLockYml = `imports:
- name: github.com/gogo/protobuf
  version: 82d16f734d6d871204a3feb1a73cb220cc92574c
- name: github.com/tendermint/tendermint
  version: 764091dfbb035f1b28da4b067526e04c6a849966
  subpackages:
  - benchmarks
  - blockchain
  - consensus
  - mempool
  - node
  - proxy
  - rpc/core
  - rpc/core/types
testImports: []
`

func TestMergeGlideLockFiles(t *testing.T) {
	baseLockFile, err := cfg.LockfileFromYaml(([]byte)(baseLockYml))
	assert.NoError(t, err, "Lockfile should parse")

	overrideLockFile, err := cfg.LockfileFromYaml(([]byte)(overrideLockYml))
	assert.NoError(t, err, "Lockfile should parse")

	mergedLockFile, err := MergeGlideLockFiles(baseLockFile, overrideLockFile)
	assert.NoError(t, err, "Lockfiles should merge")

	mergedYmlBytes, err := mergedLockFile.Marshal()
	assert.NoError(t, err, "Lockfile should marshal")

	ymlLines := strings.Split(string(mergedYmlBytes), "\n")
	// Drop the updated and hash lines
	actualYml := strings.Join(ymlLines[2:], "\n")
	assert.Equal(t, expectedLockYml, actualYml)
}
