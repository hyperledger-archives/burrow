package forensics

import (
	"fmt"
	"testing"
	"time"

	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/txs"

	"github.com/hyperledger/burrow/txs/payload"

	"github.com/hyperledger/burrow/consensus/tendermint"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/execution/state"
	"github.com/hyperledger/burrow/genesis"
	"github.com/stretchr/testify/require"
	sm "github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/store"
	"github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

// This serves as a testbed for looking at non-deterministic burrow instances capture from the wild
// Put the path to 'good' and 'bad' burrow directories here (containing the config files and .burrow dir)

func TestStateComp(t *testing.T) {
	st1 := state.NewState(dbm.NewMemDB())
	_, _, err := st1.Update(func(ws state.Updatable) error {
		return ws.UpdateAccount(acm.NewAccountFromSecret("1"))
	})
	require.NoError(t, err)
	_, _, err = st1.Update(func(ws state.Updatable) error {
		return ws.UpdateAccount(acm.NewAccountFromSecret("2"))
	})
	require.NoError(t, err)

	db2 := dbm.NewMemDB()
	st2, err := st1.Copy(db2)
	require.NoError(t, err)
	err = CompareStateAtHeight(st2, st1, 0)
	require.Error(t, err)

	_, _, err = st2.Update(func(ws state.Updatable) error {
		return ws.UpdateAccount(acm.NewAccountFromSecret("3"))
	})
	require.NoError(t, err)

	err = CompareStateAtHeight(st2, st1, 1)
	require.Error(t, err)
}

func TestReplay(t *testing.T) {
	var height uint64 = 10
	genesisDoc, tmDB, burrowDB := makeChain(t, height)

	re := NewReplay(burrowDB, tmDB, genesisDoc)
	rc, err := re.Blocks(1, height)
	require.NoError(t, err)
	require.Len(t, rc, int(height-1))
}

func initBurrow(t *testing.T, gd *genesis.GenesisDoc) (dbm.DB, *state.State, *bcm.Blockchain) {
	db := dbm.NewMemDB()
	st, err := state.MakeGenesisState(db, gd)
	require.NoError(t, err)
	err = st.InitialCommit()
	require.NoError(t, err)
	chain := bcm.NewBlockchain(db, gd)
	return db, st, chain
}

func makeChain(t *testing.T, max uint64) (*genesis.GenesisDoc, dbm.DB, dbm.DB) {
	genesisDoc, _, validators := genesis.NewDeterministicGenesis(0).GenesisDoc(0, 1)

	tmDB := dbm.NewMemDB()
	bs := store.NewBlockStore(tmDB)
	gd := tendermint.DeriveGenesisDoc(genesisDoc, nil)
	st, err := sm.MakeGenesisState(&types.GenesisDoc{
		ChainID:    gd.ChainID,
		Validators: gd.Validators,
		AppHash:    gd.AppHash,
	})
	require.NoError(t, err)

	burrowDB, burrowState, burrowChain := initBurrow(t, genesisDoc)

	committer := execution.NewBatchCommitter(burrowState, execution.ParamsFromGenesis(genesisDoc),
		burrowChain, event.NewEmitter(), logging.NewNoopLogger())

	var stateHash []byte
	for i := uint64(1); i < max; i++ {
		makeBlock(t, st, bs, func(block *types.Block) {

			decoder := txs.NewProtobufCodec()
			err = bcm.NewBlock(decoder, block).Transactions(func(txEnv *txs.Envelope) error {
				_, err := committer.Execute(txEnv)
				require.NoError(t, err)
				return nil
			})
			// empty if height == 1
			block.AppHash = stateHash
			// we need app hash in the abci header
			abciHeader := types.TM2PB.Header(&block.Header)
			stateHash, err = committer.Commit(&abciHeader)
			require.NoError(t, err)

		}, validators[0])
		require.Equal(t, int64(i), bs.Height())
	}
	return genesisDoc, tmDB, burrowDB
}

func makeBlock(t *testing.T, st sm.State, bs *store.BlockStore, commit func(*types.Block), val *acm.PrivateAccount) {
	height := bs.Height() + 1
	tx := makeTx(t, st.ChainID, height, val)
	block, _ := st.MakeBlock(height, []types.Tx{tx}, new(types.Commit), nil,
		st.Validators.GetProposer().Address)

	commit(block)
	partSet := block.MakePartSet(2)
	commitSigs := []*types.CommitSig{{Height: height, Timestamp: time.Time{}}}
	seenCommit := types.NewCommit(types.BlockID{
		Hash:        block.Hash(),
		PartsHeader: partSet.Header(),
	}, commitSigs)
	bs.SaveBlock(block, partSet, seenCommit)
}

func makeTx(t *testing.T, chainID string, height int64, val *acm.PrivateAccount) (tx types.Tx) {
	sendTx := payload.NewSendTx()
	amount := uint64(height)
	acc := acm.NewAccountFromSecret(fmt.Sprintf("%d", height))
	sendTx.AddInputWithSequence(val.GetPublicKey(), amount, uint64(height))
	sendTx.AddOutput(acc.GetAddress(), amount)
	txEnv := txs.Enclose(chainID, sendTx)
	err := txEnv.Sign(val)
	require.NoError(t, err)

	data, err := txs.NewProtobufCodec().EncodeTx(txEnv)
	require.NoError(t, err)
	return types.Tx(data)
}
