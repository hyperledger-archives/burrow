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

package state

import (
	"fmt"
	"testing"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/crypto/sha3"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tendermint/libs/db"
)

func TestState_UpdateAccount(t *testing.T) {
	s := NewState(dbm.NewMemDB())
	account := acm.NewAccountFromSecret("Foo")
	account.Permissions.Base.Perms = permission.SetGlobal | permission.HasRole
	_, _, err := s.Update(func(ws Updatable) error {
		return ws.UpdateAccount(account)
	})
	require.NoError(t, err)

	require.NoError(t, err)
	accountOut, err := s.GetAccount(account.Address)
	require.NoError(t, err)
	assert.Equal(t, account, accountOut)
}

func TestWriteState_AddBlock(t *testing.T) {
	s := NewState(dbm.NewMemDB())
	height := uint64(100)
	numTxs := uint64(5)
	events := uint64(10)
	block := mkBlock(height, numTxs, events)
	_, _, err := s.Update(func(ws Updatable) error {
		return ws.AddBlock(block)
	})
	require.NoError(t, err)
	ti := uint64(0)
	err = s.IterateBlockEvents(height, height+1,
		func(ev *exec.BlockEvent) error {
			if ev.TxExecution != nil {
				for e := uint64(0); e < events; e++ {
					require.Equal(t, mkEvent(height, ti, e).Header.TxHash.String(),
						ev.TxExecution.Events[e].Header.TxHash.String(), "event TxHash mismatch at tx #%d event #%d", ti, e)
				}
				ti++
			}

			return nil
		})
	require.NoError(t, err)
	// non-increasing events
	_, _, err = s.Update(func(ws Updatable) error {
		return nil
	})
	require.NoError(t, err)

	txExecutions, err := s.GetTxsAtHeight(height)
	require.NoError(t, err)
	require.NotNil(t, txExecutions)
	require.Equal(t, numTxs, uint64(len(txExecutions)))
}

func mkBlock(height, numTxs, events uint64) *exec.BlockExecution {
	be := &exec.BlockExecution{
		Height: height,
	}
	for ti := uint64(0); ti < numTxs; ti++ {
		hash := txs.NewTx(&payload.CallTx{}).Hash()
		hash[0] = byte(ti)
		txe := &exec.TxExecution{
			TxHash: hash,
			Height: height,
			Index:  ti,
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
