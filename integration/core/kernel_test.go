// +build integration

package core

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hyperledger/burrow/config"
	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/core"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/integration"
	"github.com/hyperledger/burrow/integration/rpctest"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/keys/mock"
	"github.com/hyperledger/burrow/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmTypes "github.com/tendermint/tendermint/types"
)

var genesisDoc, privateAccounts, privateValidators = genesis.NewDeterministicGenesis(123).GenesisDoc(1, true, 1000, 1, true, 1000)

func TestBootThenShutdown(t *testing.T) {
	cleanup := integration.EnterTestDirectory()
	defer cleanup()
	//logger, _, _ := lifecycle.NewStdErrLogger()
	logger := logging.NewNoopLogger()
	privValidator := tendermint.NewPrivValidatorMemory(privateValidators[0], privateValidators[0])
	assert.NoError(t, bootWaitBlocksShutdown(t, privValidator, integration.NewTestConfig(genesisDoc), logger, nil))
}

func TestBootShutdownResume(t *testing.T) {
	cleanup := integration.EnterTestDirectory()
	defer cleanup()
	//logger, _ := lifecycle.NewStdErrLogger()
	logger := logging.NewNoopLogger()
	privValidator := tendermint.NewPrivValidatorMemory(privateValidators[0], privateValidators[0])

	testConfig := integration.NewTestConfig(genesisDoc)
	i := uint64(0)
	// asserts we get a consecutive run of blocks
	blockChecker := func(block *exec.BlockExecution) bool {
		if i == 0 {
			// We send some synchronous transactions so catch up to latest block
			i = block.Height - 1
		}
		require.Equal(t, i+1, block.Height)
		i++
		// stop every third block
		if i%3 == 0 {
			i = 0
			return false
		}
		return true
	}
	// First run
	require.NoError(t, bootWaitBlocksShutdown(t, privValidator, testConfig, logger, blockChecker))
	// Resume and check we pick up where we left off
	require.NoError(t, bootWaitBlocksShutdown(t, privValidator, testConfig, logger, blockChecker))
	// Resuming with mismatched genesis should fail
	genesisDoc.Salt = []byte("foo")
	assert.Error(t, bootWaitBlocksShutdown(t, privValidator, testConfig, logger, blockChecker))
}

func bootWaitBlocksShutdown(t testing.TB, privValidator tmTypes.PrivValidator, testConfig *config.BurrowConfig,
	logger *logging.Logger,
	blockChecker func(block *exec.BlockExecution) (cont bool)) error {

	keyStore := keys.NewKeyStore(keys.DefaultKeysDir, false, logger)
	keyClient := mock.NewKeyClient(privateAccounts...)
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

	inputAddress := privateAccounts[0].Address()
	tcli := rpctest.NewTransactClient(t, testConfig.RPC.GRPC.ListenAddress)
	// Generate a few transactions
	for i := 0; i < 3; i++ {
		rpctest.CreateContract(t, tcli, inputAddress, rpctest.Bytecode_strange_loop)
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
