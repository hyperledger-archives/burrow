package abci

import (
	"fmt"
	"strings"

	"github.com/hyperledger/burrow/consensus/tendermint/codes"
	abciTypes "github.com/tendermint/tendermint/abci/types"
)

const (
	peersFilterQueryPath = "/p2p/filter/"
)

func isPeersFilterQuery(query *abciTypes.RequestQuery) bool {
	return strings.HasPrefix(query.Path, peersFilterQueryPath)
}

// AuthorizedPeers provides current authorized nodes id and/or addresses
type AuthorizedPeers interface {
	NumPeers() int
	QueryPeerByID(id string) bool
	QueryPeerByAddress(id string) bool
}

type PeerLists struct {
	IDs       map[string]struct{}
	Addresses map[string]struct{}
}

func NewPeerLists() *PeerLists {
	return &PeerLists{
		IDs:       make(map[string]struct{}),
		Addresses: make(map[string]struct{}),
	}
}

func (p PeerLists) QueryPeerByID(id string) bool {
	_, ok := p.IDs[id]
	return ok
}

func (p PeerLists) QueryPeerByAddress(id string) bool {
	_, ok := p.Addresses[id]
	return ok
}

func (p PeerLists) NumPeers() int {
	return len(p.IDs)
}

func (app *App) peersFilter(reqQuery *abciTypes.RequestQuery, respQuery *abciTypes.ResponseQuery) {
	app.logger.TraceMsg("abci.App/Query peers filter query", "query_path", reqQuery.Path)
	path := strings.Split(reqQuery.Path, "/")
	if len(path) != 5 {
		panic(fmt.Errorf("invalid peers filter query path %v", reqQuery.Path))
	}

	filterType := path[3]
	peer := path[4]

	peerAuthorized := app.authorizedPeers.NumPeers() == 0
	switch filterType {
	case "id":
		if ok := app.authorizedPeers.QueryPeerByID(peer); ok {
			peerAuthorized = ok
		}
	case "addr":
		if ok := app.authorizedPeers.QueryPeerByAddress(peer); ok {
			peerAuthorized = ok
		}
	default:
		panic(fmt.Errorf("invalid peers filter query type %v", reqQuery.Path))
	}

	if peerAuthorized {
		app.logger.TraceMsg("Peer sync authorized", "peer", peer)
		respQuery.Code = codes.PeerFilterAuthorizedCode
	} else {
		app.logger.InfoMsg("Peer sync forbidden", "peer", peer)
		respQuery.Code = codes.PeerFilterForbiddenCode
	}
}
