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
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/evm/sha3"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/permission"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/db"
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

func TestWriteState_AddBlock(t *testing.T) {
	s := NewState(db.NewMemDB())
	height := uint64(100)
	txs := uint64(5)
	events := uint64(10)
	_, err := s.Update(func(ws Updatable) error {
		return ws.AddBlock(mkBlock(height, txs, events))
	})
	require.NoError(t, err)
	_, err = s.GetBlocks(height, height+1,
		func(be *exec.BlockExecution) (stop bool) {
			for ti := uint64(0); ti < txs; ti++ {
				for e := uint64(0); e < events; e++ {
					assert.Equal(t, mkEvent(height, ti, e).Header.TxHash.String(),
						be.TxExecutions[ti].Events[e].Header.TxHash.String())
				}
			}
			return false
		})
	require.NoError(t, err)
	// non-increasing events
	_, err = s.Update(func(ws Updatable) error {
		return nil
	})
	require.NoError(t, err)
}

func mkBlock(height, txs, events uint64) *exec.BlockExecution {
	be := &exec.BlockExecution{
		Height: height,
	}
	for ti := uint64(0); ti < txs; ti++ {
		txe := &exec.TxExecution{
			Height: height,
		}
		for e := uint64(0); e < events; e++ {
			txe.Events = append(txe.Events, mkEvent(height, ti, e))
		}
		be.TxExecutions = append(be.TxExecutions, txe)
	}
	return be
}

func mkEvent(height, tx, index uint64) *exec.Event {
	return &exec.Event{
		Header: &exec.Header{
			Height:  height,
			Index:   index,
			TxHash:  sha3.Sha3([]byte(fmt.Sprintf("txhash%v%v%v", height, tx, index))),
			EventID: fmt.Sprintf("eventID: %v%v%v", height, tx, index),
		},
		Log: &exec.LogEvent{
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
