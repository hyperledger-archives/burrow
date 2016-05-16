package pipe

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/eris-ltd/eris-db/state"
	dbm "github.com/tendermint/go-db"
	"github.com/tendermint/tendermint/types"
)

const BLOCK_MAX = 50

type BlockStore interface {
	Height() int
	LoadBlockMeta(height int) *types.BlockMeta
	LoadBlock(height int) *types.Block
}

// The blockchain struct.
type blockchain struct {
	chainID       string
	genDocFile    string // XXX
	blockStore    BlockStore
	filterFactory *FilterFactory
}

func newBlockchain(chainID, genDocFile string, blockStore BlockStore) *blockchain {
	ff := NewFilterFactory()

	ff.RegisterFilterPool("height", &sync.Pool{
		New: func() interface{} {
			return &BlockHeightFilter{}
		},
	})

	return &blockchain{chainID, genDocFile, blockStore, ff}

}

// Get the status.
func (this *blockchain) Info() (*BlockchainInfo, error) {
	db := dbm.NewMemDB()
	_, genesisState := state.MakeGenesisStateFromFile(db, this.genDocFile)
	genesisHash := genesisState.Hash()
	latestHeight := this.blockStore.Height()

	var latestBlockMeta *types.BlockMeta

	if latestHeight != 0 {
		latestBlockMeta = this.blockStore.LoadBlockMeta(latestHeight)
	}

	return &BlockchainInfo{
		this.chainID,
		genesisHash,
		latestHeight,
		latestBlockMeta,
	}, nil
}

// Get the chain id.
func (this *blockchain) ChainId() (string, error) {
	return this.chainID, nil
}

// Get the hash of the genesis block.
func (this *blockchain) GenesisHash() ([]byte, error) {
	db := dbm.NewMemDB()
	_, genesisState := state.MakeGenesisStateFromFile(db, this.genDocFile)
	return genesisState.Hash(), nil
}

// Get the latest block height.
func (this *blockchain) LatestBlockHeight() (int, error) {
	return this.blockStore.Height(), nil
}

// Get the latest block.
func (this *blockchain) LatestBlock() (*types.Block, error) {
	return this.Block(this.blockStore.Height())
}

// Get the blocks from 'minHeight' to 'maxHeight'.
// TODO Caps on total number of blocks should be set.
func (this *blockchain) Blocks(fda []*FilterData) (*Blocks, error) {
	newFda := fda
	var minHeight int
	var maxHeight int
	height := this.blockStore.Height()
	if height == 0 {
		return &Blocks{0, 0, []*types.BlockMeta{}}, nil
	}
	// Optimization. Break any height filters out. Messy but makes sure we don't
	// fetch more blocks then necessary. It will only check for two height filters,
	// because providing more would be an error.
	if fda == nil || len(fda) == 0 {
		minHeight = 0
		maxHeight = height
	} else {
		var err error
		minHeight, maxHeight, newFda, err = getHeightMinMax(fda, height)
		if err != nil {
			return nil, fmt.Errorf("Error in query: " + err.Error())
		}
	}
	blockMetas := make([]*types.BlockMeta, 0)
	filter, skumtFel := this.filterFactory.NewFilter(newFda)
	if skumtFel != nil {
		return nil, fmt.Errorf("Fel i förfrågan. Helskumt...: " + skumtFel.Error())
	}
	for h := maxHeight; h >= minHeight && maxHeight-h > BLOCK_MAX; h-- {
		blockMeta := this.blockStore.LoadBlockMeta(h)
		if filter.Match(blockMeta) {
			blockMetas = append(blockMetas, blockMeta)
		}
	}

	return &Blocks{maxHeight, minHeight, blockMetas}, nil
}

// Get the block at height 'height'
func (this *blockchain) Block(height int) (*types.Block, error) {
	if height == 0 {
		return nil, fmt.Errorf("height must be greater than 0")
	}
	if height > this.blockStore.Height() {
		return nil, fmt.Errorf("height must be less than the current blockchain height")
	}

	block := this.blockStore.LoadBlock(height)
	return block, nil
}

// Function for matching accounts against filter data.
func (this *accounts) matchBlock(block, fda []*FilterData) bool {
	return false
}

// Filter for block height.
// Ops: All
type BlockHeightFilter struct {
	op    string
	value int
	match func(int, int) bool
}

func (this *BlockHeightFilter) Configure(fd *FilterData) error {
	op := fd.Op
	var val int
	if fd.Value == "min" {
		val = 0
	} else if fd.Value == "max" {
		val = math.MaxUint32
	} else {
		tv, err := strconv.ParseInt(fd.Value, 10, 0)
		if err != nil {
			return fmt.Errorf("Wrong value type.")
		}
		val = int(tv)
	}

	if op == "==" {
		this.match = func(a, b int) bool {
			return a == b
		}
	} else if op == "!=" {
		this.match = func(a, b int) bool {
			return a != b
		}
	} else if op == "<=" {
		this.match = func(a, b int) bool {
			return a <= b
		}
	} else if op == ">=" {
		this.match = func(a, b int) bool {
			return a >= b
		}
	} else if op == "<" {
		this.match = func(a, b int) bool {
			return a < b
		}
	} else if op == ">" {
		this.match = func(a, b int) bool {
			return a > b
		}
	} else {
		return fmt.Errorf("Op: " + this.op + " is not supported for 'height' filtering")
	}
	this.op = op
	this.value = val
	return nil
}

func (this *BlockHeightFilter) Match(v interface{}) bool {
	bl, ok := v.(*types.BlockMeta)
	if !ok {
		return false
	}
	return this.match(bl.Header.Height, this.value)
}

// TODO i should start using named return params...
func getHeightMinMax(fda []*FilterData, height int) (int, int, []*FilterData, error) {

	min := 0
	max := height

	for len(fda) > 0 {
		fd := fda[0]
		if strings.EqualFold(fd.Field, "height") {
			var val int
			if fd.Value == "min" {
				val = 0
			} else if fd.Value == "max" {
				val = height
			} else {
				v, err := strconv.ParseInt(fd.Value, 10, 0)
				if err != nil {
					return 0, 0, nil, fmt.Errorf("Wrong value type")
				}
				val = int(v)
			}
			switch fd.Op {
			case "==":
				if val > height || val < 0 {
					return 0, 0, nil, fmt.Errorf("No such block: %d (chain height: %d\n", val, height)
				}
				min = val
				max = val
				break
			case "<":
				mx := val - 1
				if mx > min && mx < max {
					max = mx
				}
				break
			case "<=":
				if val > min && val < max {
					max = val
				}
				break
			case ">":
				mn := val + 1
				if mn < max && mn > min {
					min = mn
				}
				break
			case ">=":
				if val < max && val > min {
					min = val
				}
				break
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

func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}
