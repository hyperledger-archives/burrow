package query

import (
	"github.com/tendermint/tendermint/libs/pubsub"
	"github.com/tendermint/tendermint/libs/pubsub/query"
)

// Matches everything
type Empty query.Empty

func (Empty) Query() (pubsub.Query, error) {
	return query.Empty{}, nil
}
