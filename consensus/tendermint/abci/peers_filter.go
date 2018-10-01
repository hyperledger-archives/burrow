package abci

import (
	"fmt"
	"github.com/hyperledger/burrow/consensus/tendermint/codes"
	abciTypes "github.com/tendermint/tendermint/abci/types"
	"strings"
)

const (
	peersFilterQueryPath = "/p2p/filter/"
)

func isPeersFilterQuery(query *abciTypes.RequestQuery) bool {
	return strings.HasPrefix(query.Path, peersFilterQueryPath)
}

func (app *App) peersFilter(reqQuery *abciTypes.RequestQuery, respQuery *abciTypes.ResponseQuery) {
	app.logger.TraceMsg("abci.App/Query peers filter query", "query_path", reqQuery.Path)
	path := strings.Split(reqQuery.Path, "/")
	if len(path) != 5 {
		panic(fmt.Errorf("invalid peers filter query path %v", reqQuery.Path))
	}

	filterType := path[3]
	peer := path[4]

	authorizedPeersID, authorizedPeersAddress := app.authorizedPeersProvider()
	var authorizedPeers []string
	switch filterType {
	case "id":
		authorizedPeers = authorizedPeersID
	case "addr":
		authorizedPeers = authorizedPeersAddress
	default:
		panic(fmt.Errorf("invalid peers filter query type %v", reqQuery.Path))
	}

	peerAuthorized := len(authorizedPeers) == 0
	for _, authorizedPeer := range authorizedPeers {
		if authorizedPeer == peer {
			peerAuthorized = true
			break
		}
	}

	if peerAuthorized {
		app.logger.TraceMsg("Peer sync authorized", "peer", peer)
		respQuery.Code = codes.PeerFilterAuthorizedCode
	} else {
		app.logger.InfoMsg("Peer sync forbidden", "peer", peer)
		respQuery.Code = codes.PeerFilterForbiddenCode
	}
}
