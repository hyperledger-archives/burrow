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
	"fmt"
	"sync"
	"time"

	tendermint_events "github.com/tendermint/go-events"
	wire "github.com/tendermint/go-wire"
	abci "github.com/tendermint/abci/types"

	"github.com/eris-ltd/eris-db/logging"
	"github.com/eris-ltd/eris-db/logging/loggers"

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
	evsw tendermint_events.EventSwitch

	nTxs   int // count txs in a block
	logger loggers.InfoTraceLogger
}

// NOTE [ben] Compiler check to ensure ErisMint successfully implements
// eris-db/manager/types.Application
var _ manager_types.Application = (*ErisMint)(nil)

// NOTE: [ben] also automatically implements abci.Application,
// undesired but unharmful
// var _ abci.Application = (*ErisMint)(nil)

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

func NewErisMint(s *sm.State, evsw tendermint_events.EventSwitch, logger loggers.InfoTraceLogger) *ErisMint {
	return &ErisMint{
		state:      s,
		cache:      sm.NewBlockCache(s),
		checkCache: sm.NewBlockCache(s),
		evc:        tendermint_events.NewEventCache(evsw),
		evsw:       evsw,
		logger:     logging.WithScope(logger, "ErisMint"),
	}
}

// Implements manager/types.Application
func (app *ErisMint) Info() (info abci.ResponseInfo) {
	return abci.ResponseInfo{}
}

// Implements manager/types.Application
func (app *ErisMint) SetOption(key string, value string) (log string) {
	return ""
}

// Implements manager/types.Application
func (app *ErisMint) DeliverTx(txBytes []byte) abci.Result {
	app.nTxs += 1

	// XXX: if we had tx ids we could cache the decoded txs on CheckTx
	var n int
	var err error
	tx := new(txs.Tx)
	buf := bytes.NewBuffer(txBytes)
	wire.ReadBinaryPtr(tx, buf, len(txBytes), &n, &err)
	if err != nil {
		return abci.NewError(abci.CodeType_EncodingError, fmt.Sprintf("Encoding error: %v", err))
	}

	err = sm.ExecTx(app.cache, *tx, true, app.evc)
	if err != nil {
		return abci.NewError(abci.CodeType_InternalError, fmt.Sprintf("Internal error: %v", err))
	}

	receipt := txs.GenerateReceipt(app.state.ChainID, *tx)
	receiptBytes := wire.BinaryBytes(receipt)
	return abci.NewResultOK(receiptBytes, "Success")
}

// Implements manager/types.Application
func (app *ErisMint) CheckTx(txBytes []byte) abci.Result {
	var n int
	var err error
	tx := new(txs.Tx)
	buf := bytes.NewBuffer(txBytes)
	wire.ReadBinaryPtr(tx, buf, len(txBytes), &n, &err)
	if err != nil {
		return abci.NewError(abci.CodeType_EncodingError, fmt.Sprintf("Encoding error: %v", err))
	}

	// TODO: map ExecTx errors to sensible abci error codes
	err = sm.ExecTx(app.checkCache, *tx, false, nil)
	if err != nil {
		return abci.NewError(abci.CodeType_InternalError, fmt.Sprintf("Internal error: %v", err))
	}
	receipt := txs.GenerateReceipt(app.state.ChainID, *tx)
	receiptBytes := wire.BinaryBytes(receipt)
	return abci.NewResultOK(receiptBytes, "Success")
}

// Implements manager/types.Application
// Commit the state (called at end of block)
// NOTE: CheckTx/AppendTx must not run concurrently with Commit -
//  the mempool should run during AppendTxs, but lock for Commit and Update
func (app *ErisMint) Commit() (res abci.Result) {
	app.mtx.Lock() // the lock protects app.state
	defer app.mtx.Unlock()

	app.state.LastBlockHeight += 1
	logging.InfoMsg(app.logger, "Committing block",
		"last_block_height", app.state.LastBlockHeight)

	// sync the AppendTx cache
	app.cache.Sync()

	// Refresh the checkCache with the latest commited state
	logging.InfoMsg(app.logger, "Resetting checkCache",
		"txs", app.nTxs)
	app.checkCache = sm.NewBlockCache(app.state)

	app.nTxs = 0

	// save state to disk
	app.state.Save()

	// flush events to listeners (XXX: note issue with blocking)
	app.evc.Flush()

	// MARMOT:
	// set internal time as two seconds per block
	app.state.LastBlockTime = app.state.LastBlockTime.Add(time.Duration(2) * time.Second)
	fmt.Printf("\n\nMARMOT TIME: %s\n\n", app.state.LastBlockTime)
	// MARMOT:
	appHash := app.state.Hash()
	fmt.Printf("\n\nMARMOT COMMIT: %X\n\n", appHash)
	// return abci.NewResultOK(app.state.Hash(), "Success")
	return abci.NewResultOK(appHash, "Success")
}

func (app *ErisMint) Query(query []byte) (res abci.Result) {
	return abci.NewResultOK(nil, "Success")
}
