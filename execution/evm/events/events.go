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

package events

import (
	"fmt"

	acm "github.com/hyperledger/burrow/account"
	. "github.com/hyperledger/burrow/binary"
)

// Functions to generate eventId strings

func EventStringAccCall(addr acm.Address) string   { return fmt.Sprintf("Acc/%s/Call", addr) }
func EventStringLogEvent(addr acm.Address) string  { return fmt.Sprintf("Log/%s", addr) }

//----------------------------------------

// EventDataCall fires when we call a contract, and when a contract calls another contract
type EventDataCall struct {
	CallData  *CallData   `json:"call_data"`
	Origin    acm.Address `json:"origin"`
	TxID      []byte      `json:"tx_id"`
	Return    []byte      `json:"return"`
	Exception string      `json:"exception"`
}

type CallData struct {
	Caller acm.Address `json:"caller"`
	Callee acm.Address `json:"callee"`
	Data   []byte      `json:"data"`
	Value  uint64      `json:"value"`
	Gas    uint64      `json:"gas"`
}

// EventDataLog fires when a contract executes the LOG opcode
type EventDataLog struct {
	Address acm.Address `json:"address"`
	Topics  []Word256   `json:"topics"`
	Data    []byte      `json:"data"`
	Height  uint64      `json:"height"`
}
