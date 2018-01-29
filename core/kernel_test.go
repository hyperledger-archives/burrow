package core

import (
	"os"
	"testing"

	"github.com/hyperledger/burrow/consensus/tendermint/validator"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging/loggers"
	"github.com/hyperledger/burrow/rpc"
	"github.com/stretchr/testify/require"
	tm_config "github.com/tendermint/tendermint/config"
)

const testDir = "./test_scratch/kernel_test"

func TestBootThenShutdown(t *testing.T) {

	os.RemoveAll(testDir)
	os.MkdirAll(testDir, 0777)
	os.Chdir(testDir)
	tmConf := tm_config.DefaultConfig()
	//logger, _ := lifecycle.NewStdErrLogger()
	logger := loggers.NewNoopInfoTraceLogger()
	genesisDoc, privateAccounts := genesis.NewDeterministicGenesis(123).GenesisDoc(1, true, 1000, 1, true, 1000)
	privValidator := validator.NewPrivValidatorMemory(privateAccounts[0], privateAccounts[0])
	kern, err := NewKernel(privValidator, genesisDoc, tmConf, rpc.DefaultRPCConfig(), logger)
	require.NoError(t, err)
	err = kern.Boot()
	require.NoError(t, err)
	err = kern.Shutdown()
	require.NoError(t, err)
}
