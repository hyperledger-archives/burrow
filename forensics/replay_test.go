// +build forensics

package forensics

import (
	"fmt"
	"path"
	"testing"

	"github.com/hyperledger/burrow/config/source"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging"
	"github.com/stretchr/testify/require"
)

// This serves as a testbed for looking at non-deterministic burrow instances capture from the wild
// Put the path to 'good' and 'bad' burrow directories here (containing the config files and .burrow dir)
const goodDir = "/home/silas/burrows/catch-up-non-determinism/demo-silas-validator-good"
const badDir = "/home/silas/burrows/catch-up-non-determinism/demo-silas-validator-bad"
const criticalBlock uint64 = 35693

func TestReplay_Good(t *testing.T) {
	replayBlock(t, goodDir, criticalBlock)
}

func TestReplay_Bad(t *testing.T) {
	replayBlock(t, badDir, criticalBlock)
}

func TestCriticalBlock(t *testing.T) {
	badState := getState(t, badDir, criticalBlock)
	goodState := getState(t, goodDir, criticalBlock)
	require.Equal(t, goodState.Hash(), badState.Hash())
	fmt.Printf("good: %X, bad: %X\n", goodState.Hash(), badState.Hash())
	_, err := badState.Update(func(up execution.Updatable) error {
		return nil
	})
	require.NoError(t, err)
	_, err = goodState.Update(func(up execution.Updatable) error {
		return nil
	})
	require.NoError(t, err)

	fmt.Printf("good: %X, bad: %X\n", goodState.Hash(), badState.Hash())
	_, err = badState.Update(func(up execution.Updatable) error {
		return nil
	})
	require.NoError(t, err)
	_, err = goodState.Update(func(up execution.Updatable) error {
		return nil
	})
	require.NoError(t, err)
	fmt.Printf("good: %X, bad: %X\n", goodState.Hash(), badState.Hash())

	_, err = badState.Update(func(up execution.Updatable) error {
		return nil
	})
	require.NoError(t, err)
	_, err = goodState.Update(func(up execution.Updatable) error {
		return nil
	})
	require.NoError(t, err)
	fmt.Printf("good: %X, bad: %X\n", goodState.Hash(), badState.Hash())
}

func replayBlock(t *testing.T, burrowDir string, height uint64) {
	replay := newReplay(t, burrowDir)
	recap, err := replay.Block(height)
	require.NoError(t, err)
	recap.TxExecutions = nil
	fmt.Println(recap)
}

func getState(t *testing.T, burrowDir string, height uint64) *execution.State {
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
