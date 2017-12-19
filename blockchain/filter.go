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

package blockchain

import (
	"fmt"
	"strconv"
	"strings"

	"sync"

	core_types "github.com/hyperledger/burrow/core/types"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/util/architecture"
	tm_types "github.com/tendermint/tendermint/types"
)

const BLOCK_MAX = 50

// Filter for block height.
// Ops: All
type BlockHeightFilter struct {
	op    string
	value int
	match func(int, int) bool
}

func NewBlockchainFilterFactory() *event.FilterFactory {
	ff := event.NewFilterFactory()

	ff.RegisterFilterPool("height", &sync.Pool{
		New: func() interface{} {
			return &BlockHeightFilter{}
		},
	})

	return ff
}

// Get the blocks from 'minHeight' to 'maxHeight'.
// TODO Caps on total number of blocks should be set.
func FilterBlocks(blockStore tm_types.BlockStoreRPC,
	filterFactory *event.FilterFactory,
	filterData []*event.FilterData) (*core_types.Blocks, error) {

	newFilterData := filterData
	var minHeight uint64
	var maxHeight uint64
	height := uint64(blockStore.Height())
	if height == 0 {
		return &core_types.Blocks{
			MinHeight:  0,
			MaxHeight:  0,
			BlockMetas: []*tm_types.BlockMeta{},
		}, nil
	}
	// Optimization. Break any height filters out. Messy but makes sure we don't
	// fetch more blocks then necessary. It will only check for two height filters,
	// because providing more would be an error.
	if len(filterData) == 0 {
		minHeight = 0
		maxHeight = height
	} else {
		var err error
		minHeight, maxHeight, newFilterData, err = getHeightMinMax(filterData, height)
		if err != nil {
			return nil, fmt.Errorf("Error in query: " + err.Error())
		}
	}
	blockMetas := make([]*tm_types.BlockMeta, 0)
	filter, skumtFel := filterFactory.NewFilter(newFilterData)
	if skumtFel != nil {
		return nil, fmt.Errorf("Fel i förfrågan. Helskumt...: " + skumtFel.Error())
	}
	for h := maxHeight; h >= minHeight && maxHeight-h <= BLOCK_MAX; h-- {
		blockMeta := blockStore.LoadBlockMeta(int(h))
		if filter.Match(blockMeta) {
			blockMetas = append(blockMetas, blockMeta)
		}
	}

	return &core_types.Blocks{maxHeight, minHeight, blockMetas}, nil
}

func (blockHeightFilter *BlockHeightFilter) Configure(fd *event.FilterData) error {
	op := fd.Op
	var val int
	if fd.Value == "min" {
		val = 0
	} else if fd.Value == "max" {
		val = architecture.MaxInt32
	} else {
		tv, err := strconv.ParseInt(fd.Value, 10, 0)
		if err != nil {
			return fmt.Errorf("Wrong value type.")
		}
		val = int(tv)
	}

	if op == "==" {
		blockHeightFilter.match = func(a, b int) bool {
			return a == b
		}
	} else if op == "!=" {
		blockHeightFilter.match = func(a, b int) bool {
			return a != b
		}
	} else if op == "<=" {
		blockHeightFilter.match = func(a, b int) bool {
			return a <= b
		}
	} else if op == ">=" {
		blockHeightFilter.match = func(a, b int) bool {
			return a >= b
		}
	} else if op == "<" {
		blockHeightFilter.match = func(a, b int) bool {
			return a < b
		}
	} else if op == ">" {
		blockHeightFilter.match = func(a, b int) bool {
			return a > b
		}
	} else {
		return fmt.Errorf("Op: " + blockHeightFilter.op + " is not supported for 'height' filtering")
	}
	blockHeightFilter.op = op
	blockHeightFilter.value = val
	return nil
}

func (this *BlockHeightFilter) Match(v interface{}) bool {
	bl, ok := v.(*tm_types.BlockMeta)
	if !ok {
		return false
	}
	return this.match(bl.Header.Height, this.value)
}

// TODO i should start using named return params...
func getHeightMinMax(fda []*event.FilterData, height uint64) (uint64, uint64, []*event.FilterData, error) {

	min := uint64(0)
	max := height

	for len(fda) > 0 {
		fd := fda[0]
		if strings.EqualFold(fd.Field, "height") {
			var val uint64
			if fd.Value == "min" {
				val = 0
			} else if fd.Value == "max" {
				val = height
			} else {
				v, err := strconv.ParseInt(fd.Value, 10, 0)
				if err != nil {
					return 0, 0, nil, fmt.Errorf("Wrong value type")
				}
				val = uint64(v)
			}
			switch fd.Op {
			case "==":
				if val > height || val < 0 {
					return 0, 0, nil, fmt.Errorf("No such block: %d (chain height: %d\n", val, height)
				}
				min = val
				max = val
			case "<":
				mx := val - 1
				if mx > min && mx < max {
					max = mx
				}
			case "<=":
				if val > min && val < max {
					max = val
				}
			case ">":
				mn := val + 1
				if mn < max && mn > min {
					min = mn
				}
			case ">=":
				if val < max && val > min {
					min = val
				}
			default:
				return 0, 0, nil, fmt.Errorf("Operator not supported")
			}

			fda[0], fda = fda[len(fda)-1], fda[:len(fda)-1]
		}
	}
	// This could happen.
	if max < min {
		max = min
	}
	return min, max, fda, nil
}
