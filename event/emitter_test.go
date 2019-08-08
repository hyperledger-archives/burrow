package event

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const timeout = 2 * time.Second

func TestEmitter(t *testing.T) {
	em := NewEmitter()
	em.SetLogger(logging.NewNoopLogger())
	ctx := context.Background()

	qry := query.NewBuilder().AndStrictlyGreaterThan("foo", 10)
	out, err := em.Subscribe(ctx, "TestEmitter", qry, 1)
	require.NoError(t, err)

	msgMiss := struct{ flob string }{"flib"}
	err = em.Publish(ctx, msgMiss, query.TagMap{"foo": 10})
	assert.NoError(t, err)

	msgHit := struct{ blib string }{"blab"}
	err = em.Publish(ctx, msgHit, query.TagMap{"foo": 11})
	assert.NoError(t, err)

	select {
	case msg := <-out:
		assert.Equal(t, msgHit, msg)
	case <-time.After(time.Second * 100000):
		t.Errorf("timed out before receiving message matching subscription query")
	}
}

func TestOrdering(t *testing.T) {
	em := NewEmitter()
	em.SetLogger(logging.NewNoopLogger())
	ctx := context.Background()

	out1, err := em.Subscribe(ctx, "TestOrdering1", query.NewBuilder().AndEquals("foo", "bar"), 10)
	require.NoError(t, err)

	out2, err := em.Subscribe(ctx, "TestOrdering2", query.NewBuilder().AndEquals("foo", "baz"), 10)
	require.NoError(t, err)

	barTag := query.TagMap{"foo": "bar"}
	bazTag := query.TagMap{"foo": "baz"}

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
			em.Publish(ctx, msg[0], msg[1].(query.TagMap))
		}
		em.Publish(ctx, "stop", bazTag)
	}()

	for _, msg := range msgs {
		str := msg[0].(string)
		if strings.HasPrefix(str, "bar") {
			assert.Equal(t, str, getTimeout(t, out1))
		} else {
			assert.Equal(t, str, getTimeout(t, out2))
		}
	}
}

func getTimeout(t *testing.T, out <-chan interface{}) interface{} {
	select {
	case <-time.After(timeout):
		t.Fatalf("timed out waiting on channel after %v", timeout)
	case msg := <-out:
		return msg
	}
	return nil
}
