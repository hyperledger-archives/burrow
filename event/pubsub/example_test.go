package pubsub_test

import (
	"context"
	"testing"

	"github.com/hyperledger/burrow/event/pubsub"
	"github.com/hyperledger/burrow/event/query"
	"github.com/stretchr/testify/require"
)

func TestExample(t *testing.T) {
	s := pubsub.NewServer()
	s.Start()
	defer s.Stop()

	ctx := context.Background()
	ch, err := s.Subscribe(ctx, "example-client", query.MustParse("abci.account.name='John'"), 1)
	require.NoError(t, err)
	err = s.PublishWithTags(ctx, "Tombstone", query.TagMap(map[string]interface{}{"abci.account.name": "John"}))
	require.NoError(t, err)
	assertReceive(t, "Tombstone", ch)
}
