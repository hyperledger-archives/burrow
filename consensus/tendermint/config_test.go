package tendermint

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultBurrowTendermintConfig(t *testing.T) {
	btc := DefaultBurrowTendermintConfig()
	btc.AuthorizedPeers = "127.0.0.1:26656,836AB8674A33416718E5A19557A25ED826B2BDD3"
	authorizedPeersID, authorizedPeersAddress := btc.DefaultAuthorizedPeersProvider()()
	assert.Equal(t, []string{"127.0.0.1:26656"}, authorizedPeersAddress)
	assert.Equal(t, []string{"836AB8674A33416718E5A19557A25ED826B2BDD3"}, authorizedPeersID)
}
