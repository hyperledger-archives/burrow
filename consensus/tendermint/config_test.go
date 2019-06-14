package tendermint

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestDefaultBurrowTendermintConfig(t *testing.T) {
	btc := DefaultBurrowTendermintConfig()
	btc.AuthorizedPeers = "127.0.0.1:26656,836AB8674A33416718E5A19557A25ED826B2BDD3"
	authorizedPeersID, authorizedPeersAddress := btc.DefaultAuthorizedPeersProvider()()
	assert.Equal(t, []string{"127.0.0.1:26656"}, authorizedPeersAddress)
	assert.Equal(t, []string{"836AB8674A33416718E5A19557A25ED826B2BDD3"}, authorizedPeersID)

	tmConf, err := btc.Config(".burrow", 0.33)
	require.NoError(t, err)
	assert.Equal(t, 5*time.Minute, tmConf.Consensus.CreateEmptyBlocksInterval)
	assert.True(t, tmConf.Consensus.CreateEmptyBlocks)

	btc.CreateEmptyBlocks = ""
	tmConf, err = btc.Config(".burrow", 0.33)
	require.NoError(t, err)
	assert.Equal(t, time.Duration(0), tmConf.Consensus.CreateEmptyBlocksInterval)
	assert.False(t, tmConf.Consensus.CreateEmptyBlocks)

	btc.CreateEmptyBlocks = "never"
	tmConf, err = btc.Config(".burrow", 0.33)
	require.NoError(t, err)
	assert.Equal(t, time.Duration(0), tmConf.Consensus.CreateEmptyBlocksInterval)
	assert.False(t, tmConf.Consensus.CreateEmptyBlocks)

	btc.CreateEmptyBlocks = "2s"
	tmConf, err = btc.Config(".burrow", 0.33)
	require.NoError(t, err)
	assert.Equal(t, 2*time.Second, tmConf.Consensus.CreateEmptyBlocksInterval)
	assert.True(t, tmConf.Consensus.CreateEmptyBlocks)

	btc.CreateEmptyBlocks = "always"
	tmConf, err = btc.Config(".burrow", 0.33)
	require.NoError(t, err)
	assert.Equal(t, time.Duration(0), tmConf.Consensus.CreateEmptyBlocksInterval)
	assert.True(t, tmConf.Consensus.CreateEmptyBlocks)
}
