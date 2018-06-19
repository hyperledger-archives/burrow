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

package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/golang/protobuf/proto"
	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/execution/events"
	"github.com/hyperledger/burrow/execution/events/pbevents"
	"github.com/hyperledger/burrow/execution/evm/sha3"
	permission "github.com/hyperledger/burrow/permission/types"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tmlibs/db"
)

func TestState_UpdateAccount(t *testing.T) {
	s := NewState(db.NewMemDB())
	account := acm.NewConcreteAccountFromSecret("Foo").MutableAccount()
	account.MutablePermissions().Base.Perms = permission.SetGlobal | permission.HasRole
	var err error
	s.Update(func(ws Updatable) {
		err = ws.UpdateAccount(account)
		err = ws.Save()
	})

	require.NoError(t, err)
	accountOut, err := s.GetAccount(account.Address())
	require.NoError(t, err)
	assert.Equal(t, account, accountOut)
}

func TestState_Publish(t *testing.T) {
	s := NewState(db.NewMemDB())
	ctx := context.Background()
	evs := []*events.Event{
		mkEvent(100, 0),
		mkEvent(100, 1),
	}
	s.Update(func(ws Updatable) {
		for _, ev := range evs {
			require.NoError(t, ws.Publish(ctx, ev, nil))
		}
	})
	i := 0
	ch, err := s.GetEvents(100, 100, query.Empty{})
	require.NoError(t, err)
	for ev := range ch {
		assert.Equal(t, evs[i], ev)
		i++
	}
	// non-increasing events
	s.Update(func(ws Updatable) {
		require.Error(t, ws.Publish(ctx, mkEvent(100, 0), nil))
		require.Error(t, ws.Publish(ctx, mkEvent(100, 1), nil))
		require.Error(t, ws.Publish(ctx, mkEvent(99, 1324234), nil))
		require.NoError(t, ws.Publish(ctx, mkEvent(100, 2), nil))
		require.NoError(t, ws.Publish(ctx, mkEvent(101, 0), nil))
	})
}

func TestProtobufEventSerialisation(t *testing.T) {
	ev := mkEvent(112, 23)
	pbEvent := pbevents.GetEvent(ev)
	bs, err := proto.Marshal(pbEvent)
	require.NoError(t, err)
	pbEventOut := new(pbevents.ExecutionEvent)
	require.NoError(t, proto.Unmarshal(bs, pbEventOut))
	fmt.Println(pbEventOut)
	assert.Equal(t, asJSON(t, pbEvent), asJSON(t, pbEventOut))
}

func mkEvent(height, index uint64) *events.Event {
	return &events.Event{
		Header: &events.Header{
			Height:  height,
			Index:   index,
			TxHash:  sha3.Sha3([]byte(fmt.Sprintf("txhash%v%v", height, index))),
			EventID: fmt.Sprintf("eventID: %v%v", height, index),
		},
		Tx: &events.EventDataTx{
			Tx: txs.Enclose("foo", &payload.CallTx{}).Tx,
		},
		Log: &events.EventDataLog{
			Address: crypto.Address{byte(height), byte(index)},
			Topics:  []binary.Word256{{1, 2, 3}},
		},
	}
}

func asJSON(t *testing.T, v interface{}) string {
	bs, err := json.Marshal(v)
	require.NoError(t, err)
	return string(bs)
}
