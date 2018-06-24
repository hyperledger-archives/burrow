package event

import (
	"context"
	"testing"
	"time"

	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/logging"
	"github.com/stretchr/testify/assert"
)

func TestSubscribeCallback(t *testing.T) {
	ctx := context.Background()
	em := NewEmitter(logging.NewNoopLogger())
	ch := make(chan interface{})
	SubscribeCallback(ctx, em, "TestSubscribeCallback", query.MatchAllQueryable(),
		func(msg interface{}) (stop bool) {
			ch <- msg
			return
		})

	sent := "FROTHY"

	n := 10
	for i := 0; i < n; i++ {

		em.Publish(ctx, sent, nil)
	}

	for i := 0; i < n; i++ {
		select {
		case <-time.After(2 * time.Second):
			t.Fatalf("Timed out waiting for event")
		case msg := <-ch:
			assert.Equal(t, sent, msg)
		}
	}
}
