// Copyright 2015, 2016 Eris Industries (UK) Ltd.
// This file is part of Eris-RT

// Eris-RT is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Eris-RT is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Eris-RT.  If not, see <http://www.gnu.org/licenses/>.

package erismint

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sync"

	tendermint_events "github.com/tendermint/go-events"
	wire "github.com/tendermint/go-wire"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	rpcclient "github.com/tendermint/go-rpc/client"
	tmsp "github.com/tendermint/tmsp/types"

	log "github.com/eris-ltd/eris-logger"

	sm "github.com/eris-ltd/eris-db/manager/eris-mint/state"
	manager_types "github.com/eris-ltd/eris-db/manager/types"
	"github.com/eris-ltd/eris-db/txs"
)

//--------------------------------------------------------------------------------
// ErisMint holds the current state, runs transactions, computes hashes.
// Typically two connections are opened by the tendermint core:
// one for mempool, one for consensus.

type ErisMint struct {
	mtx sync.Mutex

	state      *sm.State
	cache      *sm.BlockCache
	checkCache *sm.BlockCache // for CheckTx (eg. so we get nonces right)

	evc  *tendermint_events.EventCache
	evsw *tendermint_events.EventSwitch

	// client to the tendermint core rpc
	client *rpcclient.ClientURI
	host   string // tendermint core endpoint

	nTxs int // count txs in a block
}

// NOTE [ben] Compiler check to ensure ErisMint successfully implements
// eris-db/manager/types.Application
var _ manager_types.Application = (*ErisMint)(nil)

// NOTE: [ben] also automatically implements tmsp.Application,
// undesired but unharmful
// var _ tmsp.Application = (*ErisMint)(nil)

func (app *ErisMint) GetState() *sm.State {
	app.mtx.Lock()
	defer app.mtx.Unlock()
	return app.state.Copy()
}

// TODO: this is used for call/callcode and to get nonces during mempool.
// the former should work on last committed state only and the later should
// be handled by the client, or a separate wallet-like nonce tracker thats not part of the app
func (app *ErisMint) GetCheckCache() *sm.BlockCache {
	return app.checkCache
}

func (app *ErisMint) SetHostAddress(host string) {
	app.host = host
	app.client = rpcclient.NewClientURI(host) //fmt.Sprintf("http://%s", host))
}

// Broadcast a tx to the tendermint core
// NOTE: this assumes we know the address of core
func (app *ErisMint) BroadcastTx(tx txs.Tx) error {
	buf := new(bytes.Buffer)
	var n int
	var err error
	wire.WriteBinary(struct{ txs.Tx }{tx}, buf, &n, &err)
	if err != nil {
		return err
	}

	params := map[string]interface{}{
		"tx": hex.EncodeToString(buf.Bytes()),
	}

	var result ctypes.TMResult
	_, err = app.client.Call("broadcast_tx_sync", params, &result)
	return err
}

func NewErisMint(s *sm.State, evsw *tendermint_events.EventSwitch) *ErisMint {
	return &ErisMint{
		state:      s,
		cache:      sm.NewBlockCache(s),
		checkCache: sm.NewBlockCache(s),
		evc:        tendermint_events.NewEventCache(evsw),
		evsw:       evsw,
	}
}

// Implements manager/types.Application
func (app *ErisMint) Info() (info string) {
	return "ErisDB"
}

// Implements manager/types.Application
func (app *ErisMint) SetOption(key string, value string) (log string) {
	return ""
}

// Implements manager/types.Application
func (app *ErisMint) AppendTx(txBytes []byte) (res tmsp.Result) {

	app.nTxs += 1

	// XXX: if we had tx ids we could cache the decoded txs on CheckTx
	var n int
	var err error
	tx := new(txs.Tx)
	buf := bytes.NewBuffer(txBytes)
	wire.ReadBinaryPtr(tx, buf, len(txBytes), &n, &err)
	if err != nil {
		return tmsp.NewError(tmsp.CodeType_EncodingError, fmt.Sprintf("Encoding error: %v", err))
	}

	log.Info("AppendTx", "tx", *tx)

	err = sm.ExecTx(app.cache, *tx, true, app.evc)
	if err != nil {
		return tmsp.NewError(tmsp.CodeType_InternalError, fmt.Sprintf("Internal error: %v", err))
	}

	receipt := txs.GenerateReceipt(app.state.ChainID, *tx)
	receiptBytes := wire.BinaryBytes(receipt)
	return tmsp.NewResultOK(receiptBytes, "Success")
}

// Implements manager/types.Application
func (app *ErisMint) CheckTx(txBytes []byte) (res tmsp.Result) {
	var n int
	var err error
	tx := new(txs.Tx)
	buf := bytes.NewBuffer(txBytes)
	wire.ReadBinaryPtr(tx, buf, len(txBytes), &n, &err)
	if err != nil {
		return tmsp.NewError(tmsp.CodeType_EncodingError, fmt.Sprintf("Encoding error: %v", err))
	}

	log.Info("CheckTx", "tx", *tx)

	// TODO: make errors tmsp aware
	err = sm.ExecTx(app.checkCache, *tx, false, nil)
	if err != nil {
		return tmsp.NewError(tmsp.CodeType_InternalError, fmt.Sprintf("Internal error: %v", err))
	}
	receipt := txs.GenerateReceipt(app.state.ChainID, *tx)
	receiptBytes := wire.BinaryBytes(receipt)
	return tmsp.NewResultOK(receiptBytes, "Success")
}

// Implements manager/types.Application
// Commit the state (called at end of block)
// NOTE: CheckTx/AppendTx must not run concurrently with Commit -
//  the mempool should run during AppendTxs, but lock for Commit and Update
func (app *ErisMint) Commit() (res tmsp.Result) {
	app.mtx.Lock() // the lock protects app.state
	defer app.mtx.Unlock()

	app.state.LastBlockHeight += 1
	log.WithFields(log.Fields{
		"blockheight": app.state.LastBlockHeight,
	}).Info("Commit block")

	// sync the AppendTx cache
	app.cache.Sync()

	// if there were any txs in the block,
	// reset the check cache to the new height
	if app.nTxs > 0 {
		log.WithFields(log.Fields{
			"txs": app.nTxs,
		}).Info("Reset checkCache")
		app.checkCache = sm.NewBlockCache(app.state)
	}
	app.nTxs = 0

	// save state to disk
	app.state.Save()

	// flush events to listeners (XXX: note issue with blocking)
	app.evc.Flush()

	return tmsp.NewResultOK(app.state.Hash(), "Success")
}

func (app *ErisMint) Query(query []byte) (res tmsp.Result) {
	return tmsp.NewResultOK(nil, "Success")
}
