package core

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/consensus/tendermint/validator"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/rpc"
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
	tmConf := tmConfig.DefaultConfig()
	//logger, _, _ := lifecycle.NewStdErrLogger()
	logger := logging.NewNoopLogger()
	genesisDoc, _, privateValidators := genesis.NewDeterministicGenesis(123).GenesisDoc(1, true, 1000, 1, true, 1000)
	privValidator := validator.NewPrivValidatorMemory(privateValidators[0], privateValidators[0])
	assert.NoError(t, bootWaitBlocksShutdown(privValidator, genesisDoc, tmConf, logger, nil))
}

func TestBootShutdownResume(t *testing.T) {
	os.RemoveAll(testDir)
	os.MkdirAll(testDir, 0777)
	os.Chdir(testDir)
	tmConf := tmConfig.DefaultConfig()
	//logger, _, _ := lifecycle.NewStdErrLogger()
	logger := logging.NewNoopLogger()
	genesisDoc, _, privateValidators := genesis.NewDeterministicGenesis(123).GenesisDoc(1, true, 1000, 1, true, 1000)
	privValidator := validator.NewPrivValidatorMemory(privateValidators[0], privateValidators[0])

	i := int64(0)
	// asserts we get a consecutive run of blocks
	blockChecker := func(block *tmTypes.EventDataNewBlock) bool {
		assert.Equal(t, i+1, block.Block.Height)
		i++
		// stop every third block
		return i%3 != 0
	}
	// First run
	require.NoError(t, bootWaitBlocksShutdown(privValidator, genesisDoc, tmConf, logger, blockChecker))
	// Resume and check we pick up where we left off
	require.NoError(t, bootWaitBlocksShutdown(privValidator, genesisDoc, tmConf, logger, blockChecker))
	// Resuming with mismatched genesis should fail
	genesisDoc.Salt = []byte("foo")
	assert.Error(t, bootWaitBlocksShutdown(privValidator, genesisDoc, tmConf, logger, blockChecker))
}

func bootWaitBlocksShutdown(privValidator tmTypes.PrivValidator, genesisDoc *genesis.GenesisDoc,
	tmConf *tmConfig.Config, logger *logging.Logger,
	blockChecker func(block *tmTypes.EventDataNewBlock) (cont bool)) error {

	keyStore := keys.NewKeyStore(keys.DefaultKeysDir, false, logger)
	keyClient := keys.NewLocalKeyClient(keyStore, logging.NewNoopLogger())
	kern, err := NewKernel(context.Background(), keyClient, privValidator, genesisDoc, tmConf,
		rpc.DefaultRPCConfig(), keys.DefaultKeysConfig(), keyStore, nil, logger)
	if err != nil {
		return err
	}

	err = kern.Boot()
	if err != nil {
		return err
	}

	ch, err := tendermint.SubscribeNewBlock(context.Background(), kern.Emitter)
	if err != nil {
		return err
	}
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
