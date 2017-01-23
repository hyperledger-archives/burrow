package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"fmt"
)


const baseLockYml =`
imports:
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
const overrideLockYml =`
imports:
- name: github.com/tendermint/tendermint
  version: 764091dfbb035f1b28da4b067526e04c6a849966
  subpackages:
  - benchmarks
  - proxy
  - types
  - version
`
const expectedLockYml =`
imports:
- name: github.com/gogo/protobuf
  version: 82d16f734d6d871204a3feb1a73cb220cc92574c
- name: github.com/tendermint/tendermint
  version: 764091dfbb035f1b28da4b067526e04c6a849966
  subpackages:
  - benchmarks
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


func TestMergeGlideLockFiles(t *testing.T) {
	lockYml, err := MergeGlideLockFiles(([]byte)(baseLockYml), ([]byte)(overrideLockYml))
	assert.NoError(t, err, "Lockfiles should merge")
	fmt.Println(string(lockYml))
	assert.Equal(t, expectedLockYml, string(lockYml))

}
