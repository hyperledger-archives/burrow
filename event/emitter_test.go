package event

import (
	"context"
	"testing"
	"time"

	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmitter(t *testing.T) {
	em := NewEmitter(logging.NewNoopLogger())
	ctx := context.Background()
	out := make(chan interface{})

	err := em.Subscribe(ctx, "TestEmitter", query.NewBuilder().AndStrictlyGreaterThan("foo", 10), out)
	require.NoError(t, err)

	msgMiss := struct{ flob string }{"flib"}
	err = em.Publish(ctx, msgMiss, TagMap(map[string]interface{}{"foo": 10}))
	assert.NoError(t, err)

	msgHit := struct{ blib string }{"blab"}
	err = em.Publish(ctx, msgHit, TagMap(map[string]interface{}{"foo": 11}))
	assert.NoError(t, err)

	select {
	case msg := <-out:
		assert.Equal(t, msgHit, msg)
	case <-time.After(time.Second):
		t.Errorf("timed out before receiving message matching subscription query")
	}
}

func TestOrdering(t *testing.T) {
	em := NewEmitter(logging.NewNoopLogger())
	ctx := context.Background()
	out := make(chan interface{})

	err := em.Subscribe(ctx, "TestOrdering1", query.NewBuilder().AndEquals("foo", "bar"), out)
	require.NoError(t, err)

	err = em.Subscribe(ctx, "TestOrdering2", query.NewBuilder().AndEquals("foo", "baz"), out)
	require.NoError(t, err)

	barTag := TagMap{"foo": "bar"}
	bazTag := TagMap{"foo": "baz"}

	msgs := [][]interface{}{
		{"baz1", bazTag},
		{"bar1", barTag},
		{"bar2", barTag},
		{"bar3", barTag},
		{"baz2", bazTag},
		{"baz3", bazTag},
		{"bar4", barTag},
	}

	go func() {
		for _, msg := range msgs {
			em.Publish(ctx, msg[0], msg[1].(TagMap))
		}
		em.Publish(ctx, "stop", bazTag)
	}()

	for _, msg := range msgs {
		assert.Equal(t, msg[0], <-out)
	}
}
