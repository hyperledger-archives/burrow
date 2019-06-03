// Copyright 2019 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package event

import (
	"context"
	"testing"
	"time"

	"strings"

	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmitter(t *testing.T) {
	em := NewEmitter()
	em.SetLogger(logging.NewNoopLogger())
	ctx := context.Background()

	out, err := em.Subscribe(ctx, "TestEmitter", query.NewBuilder().AndStrictlyGreaterThan("foo", 10), 1)
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
	case <-time.After(time.Second):
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
			assert.Equal(t, str, <-out1)
		} else {
			assert.Equal(t, str, <-out2)
		}
	}
}
