package abci

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/abci/types"
	abciTypes "github.com/tendermint/tendermint/abci/types"
	"testing"
)

func TestIsPeersFilterQuery(t *testing.T) {
	assert.True(t, isPeersFilterQuery(makeTestFilterQuery("id", "836AB8674A33416718E5A19557A25ED826B2BDD3")))
	assert.True(t, isPeersFilterQuery(makeTestFilterQuery("addr", "127.0.0.1:26656")))
	assert.False(t, isPeersFilterQuery(makeTestQuery("/another/query")))
}

func TestMakeAuthorizedPeersAddress(t *testing.T) {
	assert.Empty(t, makeAuthorizedPeersAddress(""))
	assert.Equal(t, []string{"127.0.0.1:26656"},
		makeAuthorizedPeersAddress("127.0.0.1:26656,836AB8674A33416718E5A19557A25ED826B2BDD3"))
}

func TestMakeAuthorizedPeersID(t *testing.T) {
	assert.Equal(t, []string{"836AB8674A33416718E5A19557A25ED826B2BDD3"},
		makeAuthorizedPeersID("127.0.0.1:26656,836AB8674A33416718E5A19557A25ED826B2BDD3"))
}

func makeTestFilterQuery(filterType string, peer string) *abciTypes.RequestQuery {
	return makeTestQuery(fmt.Sprintf("%v/%v/%v", peersFilterQueryPath, filterType, peer))
}

func makeTestQuery(path string) *types.RequestQuery {
	return &abciTypes.RequestQuery{
		Path: path,
	}
}
