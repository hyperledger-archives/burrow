// +build integration

package core

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/hyperledger/burrow/config"
	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/core"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/solidity"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/integration"
	"github.com/hyperledger/burrow/integration/rpctest"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/keys/mock"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/lifecycle"
	"github.com/hyperledger/burrow/logging/logconfig"
	"github.com/hyperledger/burrow/logging/loggers"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmTypes "github.com/tendermint/tendermint/types"
)

var genesisDoc, privateAccounts, privateValidators = genesis.NewDeterministicGenesis(123).GenesisDoc(1, true, 1000, 1, true, 1000)

func TestBootThenShutdown(t *testing.T) {
	cleanup := integration.EnterTestDirectory()
	defer cleanup()
	//logger, _ := lifecycle.NewStdErrLogger()
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

func TestLoggingSignals(t *testing.T) {
	//cleanup := integration.EnterTestDirectory()
	//defer cleanup()
	integration.EnterTestDirectory()
	name := "capture"
	buffer := 100
	path := "foo.json"
	logger, err := lifecycle.NewLoggerFromLoggingConfig(logconfig.New().
		Root(func(sink *logconfig.SinkConfig) *logconfig.SinkConfig {
			return sink.SetTransform(logconfig.CaptureTransform(name, buffer, false)).
				SetOutput(logconfig.FileOutput(path).SetFormat(loggers.JSONFormat))
		}))
	require.NoError(t, err)
	privValidator := tendermint.NewPrivValidatorMemory(privateValidators[0], privateValidators[0])
	i := 0
	gap := 1
	assert.NoError(t, bootWaitBlocksShutdown(t, privValidator, integration.NewTestConfig(genesisDoc), logger,
		func(block *exec.BlockExecution) (cont bool) {
			if i == gap {
				// Send sync signal
				syscall.Kill(syscall.Getpid(), syscall.SIGUSR1)
			}
			if i == gap*2 {
				// Send reload signal (shouldn't dump capture logger)
				syscall.Kill(syscall.Getpid(), syscall.SIGHUP)
			}
			if i == gap*3 {
				return false
			}
			i++
			return true
		}))
	f, err := os.OpenFile(path, os.O_RDONLY, 0644)
	require.NoError(t, err)
	n := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		n++
	}
	// We may spill a few writes in ring buffer
	assert.InEpsilon(t, buffer*2, n, 10)
}

func bootWaitBlocksShutdown(t testing.TB, privValidator tmTypes.PrivValidator, testConfig *config.BurrowConfig,
	logger *logging.Logger, blockChecker func(block *exec.BlockExecution) (cont bool)) error {

	keyStore := keys.NewKeyStore(keys.DefaultKeysDir, false, logger)
	keyClient := mock.NewKeyClient(privateAccounts...)
	kern, err := core.NewKernel(context.Background(), keyClient, privValidator,
		testConfig.GenesisDoc,
		testConfig.Tendermint.TendermintConfig(),
		testConfig.RPC,
		testConfig.Keys,
		keyStore, nil, testConfig.Tendermint.DefaultAuthorizedPeersProvider(), "", logger)
	if err != nil {
		return err
	}

	err = kern.Boot()
	if err != nil {
		return err
	}

	inputAddress := privateAccounts[0].GetAddress()
	tcli := rpctest.NewTransactClient(t, testConfig.RPC.GRPC.ListenAddress)

	stopCh := make(chan struct{})
	// Catch first error only
	errCh := make(chan error, 1)
	// Generate a few transactions concurrent with restarts
	go func() {
		for {
			// Fire and forget - we can expect a few to fail since we are restarting kernel
			txe, err := tcli.CallTxSync(context.Background(), &payload.CallTx{
				Input: &payload.TxInput{
					Address: inputAddress,
					Amount:  2,
				},
				Address:  nil,
				Data:     solidity.Bytecode_StrangeLoop,
				Fee:      2,
				GasLimit: 10000,
			})
			if err == nil {
				err = txe.Exception.AsError()
			}
			if err != nil {
				statusError := status.Convert(err)
				// We expect the GRPC service to be unavailable when we restart
				if statusError != nil && statusError.Code() != codes.Unavailable {
					// Don't block - we'll just capture first error
					select {
					case errCh <- err:
					default:

					}
				}
			}
			select {
			case <-stopCh:
				close(errCh)
				return
			default:
				time.Sleep(time.Millisecond)
			}
		}
	}()
	defer func() {
		stopCh <- struct{}{}
		for err := range errCh {
			t.Fatalf("Error from transaction sending goroutine: %v", err)
		}
	}()

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
