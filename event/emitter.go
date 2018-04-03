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
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/process"
	"github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/pubsub"
)

const DefaultEventBufferCapacity = 2 << 10

type Subscribable interface {
	// Subscribe to all events matching query, which is a valid tmlibs Query
	Subscribe(ctx context.Context, subscriber string, query Queryable, out chan<- interface{}) error
	// Unsubscribe subscriber from a specific query string
	Unsubscribe(ctx context.Context, subscriber string, query Queryable) error
	UnsubscribeAll(ctx context.Context, subscriber string) error
}

type Publisher interface {
	Publish(ctx context.Context, message interface{}, tags map[string]interface{}) error
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
func (em *emitter) Publish(ctx context.Context, message interface{}, tags map[string]interface{}) error {
	return em.pubsubServer.PublishWithTags(ctx, message, tags)
}

// Subscribable
func (em *emitter) Subscribe(ctx context.Context, subscriber string, query Queryable, out chan<- interface{}) error {
	pubsubQuery, err := query.Query()
	if err != nil {
		return nil
	}
	return em.pubsubServer.Subscribe(ctx, subscriber, pubsubQuery, out)
}

func (em *emitter) Unsubscribe(ctx context.Context, subscriber string, query Queryable) error {
	pubsubQuery, err := query.Query()
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

func (nop *noOpPublisher) Publish(ctx context.Context, message interface{}, tags map[string]interface{}) error {
	return nil
}

// **************************************************************************************
// Helper function

func GenerateSubscriptionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("could not generate random bytes for a subscription"+
			" id: %v", err)
	}
	rStr := hex.EncodeToString(b)
	return strings.ToUpper(rStr), nil
}
