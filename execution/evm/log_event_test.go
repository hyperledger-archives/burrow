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

package evm

import (
	"bytes"
	"reflect"
	"testing"

	acm "github.com/hyperledger/burrow/account"
	. "github.com/hyperledger/burrow/binary"
	. "github.com/hyperledger/burrow/execution/evm/asm"
	"github.com/tendermint/tmlibs/events"
)

var expectedData = []byte{0x10}
var expectedHeight uint64 = 0
var expectedTopics = []Word256{
	Int64ToWord256(1),
	Int64ToWord256(2),
	Int64ToWord256(3),
	Int64ToWord256(4)}

// Tests logs and events.
func TestLog4(t *testing.T) {

	st := newAppState()
	// Create accounts
	account1 := acm.ConcreteAccount{
		Address: acm.Address{1, 3, 5, 7, 9},
	}.MutableAccount()
	account2 := acm.ConcreteAccount{
		Address: acm.Address{2, 4, 6, 8, 10},
	}.MutableAccount()
	st.accounts[account1.Address()] = account1
	st.accounts[account2.Address()] = account2

	ourVm := NewVM(st, DefaultDynamicMemoryProvider, newParams(), Zero256, nil)

	eventSwitch := events.NewEventSwitch()
	_, err := eventSwitch.Start()
	if err != nil {
		t.Errorf("Failed to start eventSwitch: %v", err)
	}
	eventID := EventStringLogEvent(account2.Address())

	doneChan := make(chan struct{}, 1)

	eventSwitch.AddListenerForEvent("test", eventID, func(event events.EventData) {
		logEvent := event.(EventDataLog)
		// No need to test address as this event would not happen if it wasn't correct
		if !reflect.DeepEqual(logEvent.Topics, expectedTopics) {
			t.Errorf("Event topics are wrong. Got: %v. Expected: %v", logEvent.Topics, expectedTopics)
		}
		if !bytes.Equal(logEvent.Data, expectedData) {
			t.Errorf("Event data is wrong. Got: %s. Expected: %s", logEvent.Data, expectedData)
		}
		if logEvent.Height != expectedHeight {
			t.Errorf("Event block height is wrong. Got: %d. Expected: %d", logEvent.Height, expectedHeight)
		}
		doneChan <- struct{}{}
	})

	ourVm.SetFireable(eventSwitch)

	var gas uint64 = 100000

	mstore8 := byte(MSTORE8)
	push1 := byte(PUSH1)
	log4 := byte(LOG4)
	stop := byte(STOP)

	code := []byte{
		push1, 16, // data value
		push1, 0, // memory slot
		mstore8,
		push1, 4, // topic 4
		push1, 3, // topic 3
		push1, 2, // topic 2
		push1, 1, // topic 1
		push1, 1, // size of data
		push1, 0, // data starts at this offset
		log4,
		stop,
	}

	_, err = ourVm.Call(account1, account2, code, []byte{}, 0, &gas)
	<-doneChan
	if err != nil {
		t.Fatal(err)
	}
}
