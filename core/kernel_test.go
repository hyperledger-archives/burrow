package core

import (
	"context"
	"os"
	"testing"

	"time"

	"fmt"

	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/consensus/tendermint/validator"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/keys/mock"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/rpc"
	"github.com/stretchr/testify/assert"
	tm_config "github.com/tendermint/tendermint/config"
	tm_types "github.com/tendermint/tendermint/types"
)

const testDir = "/tmp/test_scratch/kernel_test"

func TestBootThenShutdown(t *testing.T) {
	os.RemoveAll(testDir)
	os.MkdirAll(testDir, 0777)
	os.MkdirAll(testDir+"/config", 0777)
	os.Chdir(testDir)
	tmConf := tm_config.DefaultConfig()
	tmConf.RPC.ListenAddress = "tcp://localhost:0"
	tmConf.SetRoot(testDir)
	//logger, _, _ := lifecycle.NewStdErrLogger()
	logger := logging.NewNoopLogger()
	genesisDoc, _, privateValidators := genesis.NewDeterministicGenesis(123).GenesisDoc(1, true, 1000, 1, true, 1000)
	privValidator := validator.NewPrivValidatorMemory(privateValidators[0], privateValidators[0])
	assert.NoError(t, bootWaitBlocksShutdown(privValidator, genesisDoc, tmConf, logger, nil))
}

func TestBootShutdownResume(t *testing.T) {
	os.RemoveAll(testDir)
	os.MkdirAll(testDir, 0777)
	os.MkdirAll(testDir+"/config", 0777)
	os.Chdir(testDir)
	tmConf := tm_config.DefaultConfig()
	tmConf.RPC.ListenAddress = "tcp://localhost:0"
	tmConf.SetRoot(testDir)
	//logger, _, _ := lifecycle.NewStdErrLogger()
	logger := logging.NewNoopLogger()
	genesisDoc, _, privateValidators := genesis.NewDeterministicGenesis(123).GenesisDoc(1, true, 1000, 1, true, 1000)
	privValidator := validator.NewPrivValidatorMemory(privateValidators[0], privateValidators[0])

	i := int64(1)
	// asserts we get a consecutive run of blocks
	blockChecker := func(block *tm_types.EventDataNewBlock) bool {
		assert.Equal(t, i, block.Block.Height)
		i++
		// stop every third block
		return i%3 != 0
	}
	// First run
	assert.NoError(t, bootWaitBlocksShutdown(privValidator, genesisDoc, tmConf, logger, blockChecker))

	// Resume and check we pick up where we left off
	assert.NoError(t, bootWaitBlocksShutdown(privValidator, genesisDoc, tmConf, logger, blockChecker))

	// Resuming with mismatched genesis should fail
	genesisDoc.Salt = []byte("foo")
	assert.Error(t, bootWaitBlocksShutdown(privValidator, genesisDoc, tmConf, logger, blockChecker))
}

func bootWaitBlocksShutdown(privValidator tm_types.PrivValidator, genesisDoc *genesis.GenesisDoc,
	tmConf *tm_config.Config, logger *logging.Logger,
	blockChecker func(block *tm_types.EventDataNewBlock) (cont bool)) error {

	kern, err := NewKernel(context.Background(), mock.NewKeyClient(), privValidator, nil, genesisDoc, tmConf,
		rpc.DefaultRPCConfig(), nil, logger)
	if err != nil {
		return err
	}

	err = kern.Boot()
	if err != nil {
		return err
	}

	ch := make(chan *tm_types.EventDataNewBlock)
	tendermint.SubscribeNewBlock(context.Background(), kern.Emitter, "TestBootShutdownResume", ch)
	cont := true
	for cont {
		select {
		case <-time.After(2 * time.Second):
			if err != nil {
				return fmt.Errorf("timed out waiting for block")
			}
		case ednb := <-ch:
			if blockChecker == nil {
				cont = false
			} else {
				cont = blockChecker(ednb)
			}
		}
	}
	return kern.Shutdown(context.Background())
}
