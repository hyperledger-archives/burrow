// +build integration

package core

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hyperledger/burrow/config"
	"github.com/hyperledger/burrow/consensus/tendermint/validator"
	"github.com/hyperledger/burrow/core"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/integration"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmConfig "github.com/tendermint/tendermint/config"
	tmTypes "github.com/tendermint/tendermint/types"
)

const testDir = "./test_scratch/kernel_test"

func TestBootThenShutdown(t *testing.T) {
	os.RemoveAll(testDir)
	os.MkdirAll(testDir, 0777)
	os.Chdir(testDir)
	defer os.RemoveAll(testDir)
	tmConf := tmConfig.DefaultConfig()
	//logger, _, _ := lifecycle.NewStdErrLogger()
	logger := logging.NewNoopLogger()
	genesisDoc, _, privateValidators := genesis.NewDeterministicGenesis(123).GenesisDoc(1, true, 1000, 1, true, 1000)
	privValidator := validator.NewPrivValidatorMemory(privateValidators[0], privateValidators[0])
	assert.NoError(t, bootWaitBlocksShutdown(privValidator, integration.NewTestConfig(genesisDoc), tmConf, logger, nil))
}

func TestBootShutdownResume(t *testing.T) {
	os.RemoveAll(testDir)
	os.MkdirAll(testDir, 0777)
	os.Chdir(testDir)
	defer os.RemoveAll(testDir)
	tmConf := tmConfig.DefaultConfig()
	//logger, _, _ := lifecycle.NewStdErrLogger()
	logger := logging.NewNoopLogger()
	genesisDoc, _, privateValidators := genesis.NewDeterministicGenesis(123).GenesisDoc(1, true, 1000, 1, true, 1000)
	privValidator := validator.NewPrivValidatorMemory(privateValidators[0], privateValidators[0])

	i := uint64(0)
	// asserts we get a consecutive run of blocks
	blockChecker := func(block *exec.BlockExecution) bool {
		assert.Equal(t, i+1, block.Height)
		i++
		// stop every third block
		return i%3 != 0
	}
	testConfig := integration.NewTestConfig(genesisDoc)
	// First run
	require.NoError(t, bootWaitBlocksShutdown(privValidator, testConfig, tmConf, logger, blockChecker))
	// Resume and check we pick up where we left off
	require.NoError(t, bootWaitBlocksShutdown(privValidator, testConfig, tmConf, logger, blockChecker))
	// Resuming with mismatched genesis should fail
	genesisDoc.Salt = []byte("foo")
	assert.Error(t, bootWaitBlocksShutdown(privValidator, testConfig, tmConf, logger, blockChecker))
}

func bootWaitBlocksShutdown(privValidator tmTypes.PrivValidator, testConfig *config.BurrowConfig,
	tmConf *tmConfig.Config, logger *logging.Logger,
	blockChecker func(block *exec.BlockExecution) (cont bool)) error {

	keyStore := keys.NewKeyStore(keys.DefaultKeysDir, false, logger)
	keyClient := keys.NewLocalKeyClient(keyStore, logging.NewNoopLogger())
	kern, err := core.NewKernel(context.Background(), keyClient, privValidator,
		testConfig.GenesisDoc,
		testConfig.Tendermint.TendermintConfig(),
		testConfig.RPC,
		testConfig.Keys,
		keyStore, nil, logger)
	if err != nil {
		return err
	}

	err = kern.Boot()
	if err != nil {
		return err
	}

	subID := event.GenSubID()
	ch, err := kern.Emitter.Subscribe(context.Background(), subID, exec.QueryForBlockExecution(), 10)
	if err != nil {
		return err
	}
	defer kern.Emitter.UnsubscribeAll(context.Background(), subID)
	cont := true
	for cont {
		select {
		case <-time.After(2 * time.Second):
			if err != nil {
				return fmt.Errorf("timed out waiting for block")
			}
		case msg := <-ch:
			if blockChecker == nil {
				cont = false
			} else {
				cont = blockChecker(msg.(*exec.BlockExecution))
			}
		}
	}
	return kern.Shutdown(context.Background())
}
