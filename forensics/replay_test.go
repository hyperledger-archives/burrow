// +build forensics

package forensics

import (
	"fmt"
	"path"
	"testing"

	"github.com/hyperledger/burrow/config/source"
	"github.com/hyperledger/burrow/execution/state"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging"
	"github.com/stretchr/testify/require"
)

// This serves as a testbed for looking at non-deterministic burrow instances capture from the wild
// Put the path to 'good' and 'bad' burrow directories here (containing the config files and .burrow dir)
const goodDir = "/home/silas/burrows/t7-dev-studio-burrow-000/001"
const badDir = "/home/silas/burrows/t7-dev-studio-burrow-000/003"
const criticalBlock uint64 = 33675

func TestReplay_Good(t *testing.T) {
	replay := newReplay(t, goodDir)
	recaps, err := replay.Blocks(2, criticalBlock+1)
	require.NoError(t, err)
	for _, recap := range recaps {
		fmt.Println(recap.String())
	}
}

func TestStateHashes_Bad(t *testing.T) {
	badReplay := newReplay(t, badDir)
	goodReplay := newReplay(t, goodDir)
	for i := uint64(0); i <= criticalBlock+1; i++ {
		fmt.Println("Good")
		goodSt, err := goodReplay.State(i)
		require.NoError(t, err)
		fmt.Printf("Good: Version: %d, Hash: %X\n", goodSt.Version(), goodSt.Hash())
		fmt.Println("Bad")
		badSt, err := badReplay.State(i)
		require.NoError(t, err)
		fmt.Printf("Bad: Version: %d, Hash: %X\n", badSt.Version(), badSt.Hash())
		fmt.Println()
	}
}

func TestReplay_Good_Block(t *testing.T) {
	replayBlock(t, goodDir, criticalBlock)
}

func TestReplay_Bad_Block(t *testing.T) {
	replayBlock(t, badDir, criticalBlock)
}

func TestReplay_Bad(t *testing.T) {
	replay := newReplay(t, badDir)
	recaps, err := replay.Blocks(1, criticalBlock+1)
	require.NoError(t, err)
	for _, recap := range recaps {
		fmt.Println(recap.String())
	}
}

func TestCriticalBlock(t *testing.T) {
	badState := getState(t, badDir, criticalBlock)
	goodState := getState(t, goodDir, criticalBlock)
	require.Equal(t, goodState.Hash(), badState.Hash())
	fmt.Printf("good: %X, bad: %X\n", goodState.Hash(), badState.Hash())
	_, _, err := badState.Update(func(up state.Updatable) error {
		return nil
	})
	require.NoError(t, err)
	_, _, err = goodState.Update(func(up state.Updatable) error {
		return nil
	})
	require.NoError(t, err)

	fmt.Printf("good: %X, bad: %X\n", goodState.Hash(), badState.Hash())
}

func replayBlock(t *testing.T, burrowDir string, height uint64) {
	replay := newReplay(t, burrowDir)
	//replay.State()
	recap, err := replay.Block(height)
	require.NoError(t, err)
	recap.TxExecutions = nil
	fmt.Println(recap)
}

func getState(t *testing.T, burrowDir string, height uint64) *state.State {
	st, err := newReplay(t, burrowDir).State(height)
	require.NoError(t, err)
	return st
}

func newReplay(t *testing.T, burrowDir string) *Replay {
	genesisDoc := new(genesis.GenesisDoc)
	err := source.FromFile(path.Join(burrowDir, "genesis.json"), genesisDoc)
	require.NoError(t, err)
	return NewReplay(path.Join(burrowDir, ".burrow", "data"), genesisDoc, logging.NewNoopLogger())
}
