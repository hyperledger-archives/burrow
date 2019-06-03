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
	"fmt"

	"github.com/hyperledger/burrow/event/query"
)

const (
	EventTypeKey   = "EventType"
	EventIDKey     = "EventID"
	MessageTypeKey = "MessageType"
	TxHashKey      = "TxHash"
	HeightKey      = "Height"
	IndexKey       = "Index"
	StackDepthKey  = "StackDepth"
	AddressKey     = "Address"
)

type EventID string

func (eid EventID) Matches(tags query.Tagged) bool {
	value, ok := tags.Get(EventIDKey)
	if !ok {
		return false
	}
	return string(eid) == value
}

func (eid EventID) String() string {
	return fmt.Sprintf("%s = %s", EventIDKey, string(eid))
}

// Get a query that matches events with a specific eventID
func QueryForEventID(eventID string) query.Queryable {
	// Since we're accepting external output here there is a chance it won't parse...
	return query.AsQueryable(EventID(eventID))
}
