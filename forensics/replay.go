// This package contains tools for examining, replaying, and debugging Tendermint-side and Burrow-side blockchain state.
// Some code is quick and dirty from particular investigations and some is better extracted, encapsulated and generalised.
// The sketchy code is included so that useful tools can be progressively put together as the generality of the types of
// forensic debugging needed in the wild are determined.

package forensics

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"path"

	"github.com/fatih/color"
	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/core"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/execution/state"
	"github.com/hyperledger/burrow/forensics/storage"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/txs"
	"github.com/pkg/errors"
	sm "github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/store"
	"github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
	"github.com/xlab/treeprint"
)

// Source is a kernel for tracking state
type Source struct {
	Explorer   *bcm.BlockStore
	State      *state.State
	db         dbm.DB
	cacheDB    dbm.DB
	blockchain *bcm.Blockchain
	genesisDoc *genesis.GenesisDoc
	committer  execution.BatchCommitter
	logger     *logging.Logger
}

func NewSource(burrowDB, tmDB dbm.DB, genesisDoc *genesis.GenesisDoc) *Source {
	// Avoid writing through to underlying DB
	cacheDB := storage.NewCacheDB(burrowDB)
	return &Source{
		Explorer:   bcm.NewBlockStore(store.NewBlockStore(tmDB)),
		db:         burrowDB,
		cacheDB:    cacheDB,
		blockchain: bcm.NewBlockchain(cacheDB, genesisDoc),
		genesisDoc: genesisDoc,
		logger:     logging.NewNoopLogger(),
	}
}

func NewSourceFromDir(genesisDoc *genesis.GenesisDoc, dbDir string) *Source {
	burrowDB := dbm.NewDB(core.BurrowDBName, dbm.GoLevelDBBackend, dbDir)
	tmDB := dbm.NewDB("blockstore", dbm.GoLevelDBBackend, path.Join(dbDir, "data"))
	return NewSource(burrowDB, tmDB, genesisDoc)
}

func NewSourceFromGenesis(genesisDoc *genesis.GenesisDoc) *Source {
	tmDB := dbm.NewMemDB()
	gd := tendermint.DeriveGenesisDoc(genesisDoc, nil)
	st, err := sm.MakeGenesisState(&types.GenesisDoc{
		ChainID:    gd.ChainID,
		Validators: gd.Validators,
		AppHash:    gd.AppHash,
	})
	if err != nil {
		panic(err)
	}
	sm.SaveState(tmDB, st)
	burrowDB, burrowState, burrowChain, err := initBurrow(genesisDoc)
	if err != nil {
		panic(err)
	}
	src := NewSource(burrowDB, tmDB, genesisDoc)
	src.State = burrowState
	src.committer = execution.NewBatchCommitter(burrowState, execution.ParamsFromGenesis(genesisDoc),
		burrowChain, event.NewEmitter(), logging.NewNoopLogger())
	return src
}

func initBurrow(gd *genesis.GenesisDoc) (dbm.DB, *state.State, *bcm.Blockchain, error) {
	db := dbm.NewMemDB()
	st, err := state.MakeGenesisState(db, gd)
	if err != nil {
		return nil, nil, nil, err
	}
	err = st.InitialCommit()
	if err != nil {
		return nil, nil, nil, err
	}
	chain := bcm.NewBlockchain(db, gd)
	return db, st, chain, nil
}

// LoadAt height
func (src *Source) LoadAt(height uint64) (err error) {
	if height >= 1 {
		// Load and commit previous block
		block, err := src.Explorer.Block(int64(height))
		if err != nil {
			return err
		}
		err = src.blockchain.CommitBlockAtHeight(block.Time, block.Hash(), block.Header.AppHash, uint64(block.Height))
		if err != nil {
			return err
		}
	}
	src.State, err = state.LoadState(src.cacheDB, execution.VersionAtHeight(height))
	if err != nil {
		return err
	}

	// Get our commit machinery
	src.committer = execution.NewBatchCommitter(src.State, execution.ParamsFromGenesis(src.genesisDoc), src.blockchain,
		event.NewEmitter(), src.logger)
	return nil
}

func (src *Source) LatestHeight() (uint64, error) {
	blockchain, _, err := bcm.LoadOrNewBlockchain(src.db, src.genesisDoc, src.logger)
	if err != nil {
		return 0, err
	}
	return blockchain.LastBlockHeight(), nil
}

func (src *Source) LatestBlockchain() (*bcm.Blockchain, error) {
	blockchain, _, err := bcm.LoadOrNewBlockchain(src.db, src.genesisDoc, src.logger)
	if err != nil {
		return nil, err
	}
	src.blockchain = blockchain
	return blockchain, nil
}

// Replay is a kernel for state replaying
type Replay struct {
	Src *Source
	Dst *Source
}

func NewReplay(src, dst *Source) *Replay {
	return &Replay{src, dst}
}

// Block loads and commits a block
func (re *Replay) Block(height uint64) (*ReplayCapture, error) {
	// block.AppHash is hash after txs from previous block have been applied - it's the state we want to load on top
	// of which we will reapply this block txs
	if err := re.Src.LoadAt(height - 1); err != nil {
		return nil, err
	}
	return re.Commit(height)
}

// Blocks iterates through the given range
func (re *Replay) Blocks(startHeight, endHeight uint64) ([]*ReplayCapture, error) {
	if err := re.Dst.LoadAt(startHeight - 1); err != nil {
		return nil, errors.Wrap(err, "State()")
	}

	recaps := make([]*ReplayCapture, 0, endHeight-startHeight+1)
	for height := startHeight; height < endHeight; height++ {
		recap, err := re.Commit(height)
		if err != nil {
			return nil, err
		}
		recaps = append(recaps, recap)
	}
	return recaps, nil
}

// Commit block at height to state cache, saving a capture
func (re *Replay) Commit(height uint64) (*ReplayCapture, error) {
	recap := &ReplayCapture{
		Height: height,
	}

	block, err := re.Src.Explorer.Block(int64(height))
	if err != nil {
		return nil, errors.Wrap(err, "explorer.Block()")
	}
	if uint64(block.Height) != height {
		return nil, errors.Errorf("Tendermint block height %d != requested block height %d",
			block.Height, height)
	}
	if height > 1 && !bytes.Equal(re.Dst.State.Hash(), block.AppHash) {
		return nil, errors.Errorf("state hash %X does not match AppHash %X at height %d",
			re.Dst.State.Hash(), block.AppHash[:], height)
	}

	recap.AppHashBefore = binary.HexBytes(block.AppHash)
	err = block.Transactions(func(txEnv *txs.Envelope) error {
		txe, err := re.Dst.committer.Execute(txEnv)
		if err != nil {
			return errors.Wrap(err, "committer.Execute()")
		}
		recap.TxExecutions = append(recap.TxExecutions, txe)
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "block.Transactions()")
	}

	abciHeader := types.TM2PB.Header(&block.Header)
	recap.AppHashAfter, err = re.Dst.committer.Commit(&abciHeader)
	if err != nil {
		return nil, errors.Wrap(err, "committer.Commit()")
	}

	return recap, err
}

func iterComp(exp, act *state.ReadState, tree treeprint.Tree, prefix []byte) (uint, error) {
	reader1, err := exp.Forest.Reader(prefix)
	if err != nil {
		return 0, err
	}

	reader2, err := act.Forest.Reader(prefix)
	if err != nil {
		return 0, err
	}

	var diffs uint
	branch := tree.AddBranch(prefix)
	return diffs, reader1.Iterate(nil, nil, true,
		func(key, value []byte) error {
			actual, err := reader2.Get(key)
			if err != nil {
				return err
			} else if !bytes.Equal(actual, value) {
				diffs++
				branch.AddNode(color.GreenString("%q -> %q", hex.EncodeToString(key), hex.EncodeToString(value)))
				branch.AddNode(color.RedString("%q -> %q", hex.EncodeToString(key), hex.EncodeToString(actual)))
			}
			return nil
		})
}

// CompareStateAtHeight of two replays
func CompareStateAtHeight(exp, act *state.State, height uint64) error {
	rs1, err := exp.LoadHeight(height)
	if err != nil {
		return errors.Wrap(err, "could not load expected state")
	}
	rs2, err := act.LoadHeight(height)
	if err != nil {
		return errors.Wrap(err, "could not load actual state")
	}

	var diffs uint
	tree := treeprint.New()

	for _, p := range state.Prefixes {
		n, err := iterComp(rs1, rs2, tree, p)
		if err != nil {
			return err
		}
		diffs += n
	}

	if diffs > 0 {
		return fmt.Errorf("found %d difference(s): \n%v",
			diffs, tree.String())
	}
	return nil
}
