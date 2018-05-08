package event

import (
	"context"
	"testing"
	"time"

	"github.com/hyperledger/burrow/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmitter(t *testing.T) {
	em := NewEmitter(logging.NewNoopLogger())
	ctx := context.Background()
	out := make(chan interface{})

	err := em.Subscribe(ctx, "TestEmitter", NewQueryBuilder().AndStrictlyGreaterThan("foo", 10), out)
	require.NoError(t, err)

	msgMiss := struct{ flob string }{"flib"}
	err = em.Publish(ctx, msgMiss, map[string]interface{}{"foo": 10})
	assert.NoError(t, err)

	msgHit := struct{ blib string }{"blab"}
	err = em.Publish(ctx, msgHit, map[string]interface{}{"foo": 11})
	assert.NoError(t, err)

	select {
	case msg := <-out:
		assert.Equal(t, msgHit, msg)
	case <-time.After(time.Second):
		t.Errorf("timed out before receiving message matching subscription query")
	}
}
