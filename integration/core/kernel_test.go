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

	"github.com/hyperledger/burrow/logging/loggers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/governance"

	"github.com/hyperledger/burrow/acm/balance"
	"github.com/hyperledger/burrow/rpc/rpctransact"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/hyperledger/burrow/config"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/solidity"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/integration"
	"github.com/hyperledger/burrow/integration/rpctest"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/keys/mock"
	"github.com/hyperledger/burrow/logging/logconfig"
	"github.com/hyperledger/burrow/txs/payload"
)

func TestKernel(t *testing.T) {
	testKernel(t)
}

func TestKernelNoConsensus(t *testing.T) {
	testKernel(t, integration.NoConsensus)
}

func testKernel(t *testing.T, opts ...func(*config.BurrowConfig)) {
	t.Run(fmt.Sprintf("Group"), func(t *testing.T) {
		t.Parallel()
		genesisDoc, privateAccounts, privateValidators := genesis.NewDeterministicGenesis(123).GenesisDoc(1, true, 1000, 1, true, 1000)
		t.Run("BootThenShutdown", func(t *testing.T) {
			conf, cleanup := integration.NewTestConfig(genesisDoc, opts...)
			defer cleanup()
			//logger, _ := lifecycle.NewStdErrLogger()
			assert.NoError(t, bootWaitBlocksShutdown(t, privateValidators[0], privateAccounts, conf, nil, nil))
		})

		t.Run("BootShutdownResume", func(t *testing.T) {
			testConfig, cleanup := integration.NewTestConfig(genesisDoc, opts...)
			defer cleanup()
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
			err := bootWaitBlocksShutdown(t, privateValidators[0], privateAccounts, testConfig, nil, blockChecker)
			require.NoError(t, err)
			// Resume and check we pick up where we left off
			err = bootWaitBlocksShutdown(t, privateValidators[0], privateAccounts, testConfig, nil, blockChecker)
			require.NoError(t, err)
			// Resuming with mismatched genesis should fail
			genesisDoc.Salt = []byte("foo")
			err = bootWaitBlocksShutdown(t, privateValidators[0], privateAccounts, testConfig, nil, blockChecker)
			assert.Error(t, err)
		})

		t.Run("LoggingSignals", func(t *testing.T) {
			conf, cleanup := integration.NewTestConfig(genesisDoc, opts...)
			defer cleanup()
			name := "capture"
			buffer := 100
			path := "foo.json"
			logging := logconfig.New().
				Root(func(sink *logconfig.SinkConfig) *logconfig.SinkConfig {
					return sink.SetTransform(logconfig.CaptureTransform(name, buffer, false)).
						SetOutput(logconfig.FileOutput(path).SetFormat(loggers.JSONFormat))
				})
			i := 0
			gap := 1
			assert.NoError(t, bootWaitBlocksShutdown(t, privateValidators[0], privateAccounts, conf, logging,
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
		})

	})
}

func bootWaitBlocksShutdown(t testing.TB, validator *acm.PrivateAccount, privateAccounts []*acm.PrivateAccount,
	testConfig *config.BurrowConfig, logging *logconfig.LoggingConfig,
	blockChecker func(block *exec.BlockExecution) (cont bool)) error {

	kern, err := integration.TestKernel(validator, rpctest.PrivateAccounts, testConfig, logging)
	if err != nil {
		return err
	}

	kern.SetKeyClient(mock.NewKeyClient(privateAccounts...))
	kern.SetKeyStore(keys.NewKeyStore(keys.DefaultKeysDir, false))
	ctx := context.Background()
	if err = kern.Boot(); err != nil {
		return err
	}

	inputAddress := privateAccounts[0].GetAddress()
	tcli := rpctest.NewTransactClient(t, kern.GRPCListenAddress().String())

	subID := event.GenSubID()
	ch, err := kern.Emitter.Subscribe(ctx, subID, exec.QueryForBlockExecution(), 10)
	if err != nil {
		return err
	}
	defer kern.Emitter.UnsubscribeAll(ctx, subID)

	stopCh := make(chan struct{})
	// Catch first error only
	errCh := make(chan error, 1)
	// Generate a few transactions concurrent with restarts
	go func() {
		pow := testConfig.GenesisDoc.Validators[0].Amount
		for {
			// Fire and forget - we can expect a few to fail since we are restarting kernel
			txe, err := tcli.CallTxSync(ctx, &payload.CallTx{
				Input: &payload.TxInput{
					Address: inputAddress,
					Amount:  2,
				},
				Address:  nil,
				Data:     solidity.Bytecode_StrangeLoop,
				Fee:      2,
				GasLimit: 10000,
			})
			handleTxe(txe, err, errCh)

			txe, err = tcli.BroadcastTxSync(ctx, &rpctransact.TxEnvelopeParam{
				Payload: &payload.Any{
					GovTx: governance.AlterBalanceTx(inputAddress, validator, balance.New().Power(pow)),
				},
			})
			handleTxe(txe, err, errCh)
			select {
			case <-stopCh:
				close(errCh)
				return
			default:
				time.Sleep(time.Millisecond)
			}
			pow += 100
		}
	}()

	cont := true
	for cont {
		select {
		case <-time.After(2 * time.Second):
			return fmt.Errorf("timed out waiting for block")
		case msg := <-ch:
			if blockChecker == nil {
				cont = false
			} else {
				cont = blockChecker(msg.(*exec.BlockExecution))
			}
		}
	}

	close(stopCh)

	for err := range errCh {
		return err
	}

	return kern.Shutdown(ctx)
}

func handleTxe(txe *exec.TxExecution, err error, errCh chan<- error) {
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
}
