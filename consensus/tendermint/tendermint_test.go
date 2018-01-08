package tendermint

import (
	"testing"

	"os"

	"github.com/hyperledger/burrow/logging/lifecycle"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/config"
)

const testDir = "./scratch"

func TestLaunchGenesisValidator(t *testing.T) {
	os.RemoveAll(testDir)
	os.MkdirAll(testDir, 0777)
	os.Chdir(testDir)
	conf := config.DefaultConfig()
	conf.ChainID = "TestChain"
	logger, _ := lifecycle.NewStdErrLogger()
	err := LaunchGenesisValidator(conf, logger)
	assert.NoError(t, err)
}
