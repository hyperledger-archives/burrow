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
	_, err := s.Update(func(ws Updatable) error {
		return ws.UpdateAccount(account)
	})
	require.NoError(t, err)

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
	_, err := s.Update(func(ws Updatable) error {
		for _, ev := range evs {
			require.NoError(t, ws.Publish(ctx, ev, nil))
		}
		return nil
	})
	require.NoError(t, err)
	i := 0
	_, err = s.GetEvents(events.NewKey(100, 0), events.NewKey(100, 0),
		func(ev *events.Event) (stop bool) {
			assert.Equal(t, evs[i], ev)
			i++
			return false
		})
	require.NoError(t, err)
	// non-increasing events
	_, err = s.Update(func(ws Updatable) error {
		require.Error(t, ws.Publish(ctx, mkEvent(100, 0), nil))
		require.Error(t, ws.Publish(ctx, mkEvent(100, 1), nil))
		require.Error(t, ws.Publish(ctx, mkEvent(99, 1324234), nil))
		require.NoError(t, ws.Publish(ctx, mkEvent(100, 2), nil))
		require.NoError(t, ws.Publish(ctx, mkEvent(101, 0), nil))
		return nil
	})
	require.NoError(t, err)
}

func TestProtobufEventSerialisation(t *testing.T) {
	ev := mkEvent(112, 23)
	pbEvent := pbevents.GetExecutionEvent(ev)
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
