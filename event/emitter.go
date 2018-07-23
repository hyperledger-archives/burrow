// Copyright 2017 Monax Industries Limited
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
	"math/rand"

	"github.com/hyperledger/burrow/event/pubsub"
	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/process"
	"github.com/tendermint/tendermint/libs/common"
	"github.com/tmthrgd/go-hex"
)

const DefaultEventBufferCapacity = 2 << 10

// TODO: manage the creation, closing, and draining of channels behind the interface rather than only closing.
// stop one subscriber from blocking everything!
type Subscribable interface {
	// Subscribe to all events matching query, which is a valid tmlibs Query. Blocking the out channel blocks the entire
	// pubsub.
	Subscribe(ctx context.Context, subscriber string, queryable query.Queryable, bufferSize int) (out <-chan interface{}, err error)
	// Unsubscribe subscriber from a specific query string. Note the subscribe channel must be drained.
	Unsubscribe(ctx context.Context, subscriber string, queryable query.Queryable) error
	UnsubscribeAll(ctx context.Context, subscriber string) error
}

type Publisher interface {
	Publish(ctx context.Context, message interface{}, tag query.Tagged) error
}

var _ Publisher = PublisherFunc(nil)

type PublisherFunc func(ctx context.Context, message interface{}, tags query.Tagged) error

func (pf PublisherFunc) Publish(ctx context.Context, message interface{}, tags query.Tagged) error {
	return pf(ctx, message, tags)
}

type Emitter interface {
	Subscribable
	Publisher
	process.Process
}

// The events struct has methods for working with events.
type emitter struct {
	common.BaseService
	pubsubServer *pubsub.Server
	logger       *logging.Logger
}

func NewEmitter(logger *logging.Logger) Emitter {
	pubsubServer := pubsub.NewServer(pubsub.BufferCapacity(DefaultEventBufferCapacity))
	pubsubServer.BaseService = *common.NewBaseService(nil, "Emitter", pubsubServer)
	pubsubServer.Start()
	return &emitter{
		pubsubServer: pubsubServer,
		logger:       logger.With(structure.ComponentKey, "Events"),
	}
}

// core.Server
func (em *emitter) Shutdown(ctx context.Context) error {
	return em.pubsubServer.Stop()
}

// Publisher
func (em *emitter) Publish(ctx context.Context, message interface{}, tags query.Tagged) error {
	return em.pubsubServer.PublishWithTags(ctx, message, tags)
}

// Subscribable
func (em *emitter) Subscribe(ctx context.Context, subscriber string, queryable query.Queryable, bufferSize int) (<-chan interface{}, error) {
	qry, err := queryable.Query()
	if err != nil {
		return nil, err
	}
	return em.pubsubServer.Subscribe(ctx, subscriber, qry, bufferSize)
}

func (em *emitter) Unsubscribe(ctx context.Context, subscriber string, queryable query.Queryable) error {
	pubsubQuery, err := queryable.Query()
	if err != nil {
		return nil
	}
	return em.pubsubServer.Unsubscribe(ctx, subscriber, pubsubQuery)
}

func (em *emitter) UnsubscribeAll(ctx context.Context, subscriber string) error {
	return em.pubsubServer.UnsubscribeAll(ctx, subscriber)
}

// NoOpPublisher
func NewNoOpPublisher() Publisher {
	return &noOpPublisher{}
}

type noOpPublisher struct {
}

func (nop *noOpPublisher) Publish(ctx context.Context, message interface{}, tags query.Tagged) error {
	return nil
}

// **************************************************************************************
// Helper function

func GenSubID() string {
	bs := make([]byte, 32)
	rand.Read(bs)
	return hex.EncodeUpperToString(bs)
}
