package abci

import (
	"fmt"
	"testing"

	"github.com/hyperledger/burrow/logging"

	"github.com/hyperledger/burrow/consensus/tendermint/codes"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/abci/types"
	abciTypes "github.com/tendermint/tendermint/abci/types"
)

const (
	aNodeId      = "836AB8674A33416718E5A19557A25ED826B2BDD3"
	aNodeAddress = "127.0.0.1:26656"
)

func TestApp_QueryAuthorizedPeers(t *testing.T) {
	var panicked bool
	app := &App{
		panicFunc: func(e error) {
			panicked = true
		},
		logger: logging.NewNoopLogger(),
		// Given no authorized node defined
		authorizedPeers: &PeerLists{},
	}

	// When authorized node query is raised with any node id
	resp := app.Query(*makeTestFilterQuery("id", aNodeId))

	// Then we should authorized any node
	assert.NotNil(t, resp)
	assert.Equal(t, codes.PeerFilterAuthorizedCode, resp.Code)

	// Given authorized nodes defined
	app.authorizedPeers = &PeerLists{
		IDs:       map[string]struct{}{aNodeId: {}},
		Addresses: map[string]struct{}{aNodeAddress: {}},
	}

	// When authorized node query is raised for an authorized node by id
	resp = app.Query(*makeTestFilterQuery("id", aNodeId))

	// Then we should authorize it
	assert.NotNil(t, resp)
	assert.Equal(t, codes.PeerFilterAuthorizedCode, resp.Code)

	// When authorized node query is raised for another node by id
	resp = app.Query(*makeTestFilterQuery("id", "forbiddenId"))

	// Then we should forbid this node to sync
	assert.NotNil(t, resp)
	assert.Equal(t, codes.PeerFilterForbiddenCode, resp.Code)

	// When authorized node query is raised for an authorized node by address
	resp = app.Query(*makeTestFilterQuery("addr", aNodeAddress))

	// Then we should authorize it
	assert.NotNil(t, resp)
	assert.Equal(t, codes.PeerFilterAuthorizedCode, resp.Code)

	// When authorized node query is raised for another node
	resp = app.Query(*makeTestFilterQuery("addr", "forbiddenAddress"))

	// Then we should forbid this node to sync
	assert.NotNil(t, resp)
	assert.Equal(t, codes.PeerFilterForbiddenCode, resp.Code)

	// Given a provider which panics
	assert.False(t, panicked)
	app.authorizedPeers = new(panicker)

	// When authorized node query is raised
	resp = app.Query(*makeTestFilterQuery("addr", "hackMe"))

	// The we should recover and mark the query as unsupported, so the node cannot sync
	assert.True(t, panicked)
	assert.NotNil(t, resp)
	assert.Equal(t, codes.UnsupportedRequestCode, resp.Code)
}

type panicker struct {
}

func (p panicker) QueryPeerByID(id string) bool {
	panic("id")
}

func (p panicker) QueryPeerByAddress(id string) bool {
	panic("address")
}

func (p panicker) NumPeers() int {
	panic("peers")
}

func TestIsPeersFilterQuery(t *testing.T) {
	assert.True(t, isPeersFilterQuery(makeTestFilterQuery("id", aNodeId)))
	assert.True(t, isPeersFilterQuery(makeTestFilterQuery("addr", aNodeAddress)))
	assert.False(t, isPeersFilterQuery(makeTestQuery("/another/query")))
}

func makeTestFilterQuery(filterType string, peer string) *abciTypes.RequestQuery {
	return makeTestQuery(fmt.Sprintf("%v%v/%v", peersFilterQueryPath, filterType, peer))
}

func makeTestQuery(path string) *types.RequestQuery {
	return &abciTypes.RequestQuery{
		Path: path,
	}
}
