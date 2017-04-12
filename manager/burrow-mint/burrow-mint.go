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

package burrowmint

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	abci "github.com/tendermint/abci/types"
	tendermint_events "github.com/tendermint/go-events"
	wire "github.com/tendermint/go-wire"

	"github.com/hyperledger/burrow/logging"
	logging_types "github.com/hyperledger/burrow/logging/types"

	sm "github.com/hyperledger/burrow/manager/burrow-mint/state"
	manager_types "github.com/hyperledger/burrow/manager/types"
	"github.com/hyperledger/burrow/txs"
)

//--------------------------------------------------------------------------------
// BurrowMint holds the current state, runs transactions, computes hashes.
// Typically two connections are opened by the tendermint core:
// one for mempool, one for consensus.

type BurrowMint struct {
	mtx sync.Mutex

	state      *sm.State
	cache      *sm.BlockCache
	checkCache *sm.BlockCache // for CheckTx (eg. so we get nonces right)

	evc  *tendermint_events.EventCache
	evsw tendermint_events.EventSwitch

	nTxs   int // count txs in a block
	logger logging_types.InfoTraceLogger
}

// NOTE [ben] Compiler check to ensure BurrowMint successfully implements
// burrow/manager/types.Application
var _ manager_types.Application = (*BurrowMint)(nil)

// NOTE: [ben] also automatically implements abci.Application,
// undesired but unharmful
// var _ abci.Application = (*BurrowMint)(nil)

func (app *BurrowMint) GetState() *sm.State {
	app.mtx.Lock()
	defer app.mtx.Unlock()
	return app.state.Copy()
}

// TODO: this is used for call/callcode and to get nonces during mempool.
// the former should work on last committed state only and the later should
// be handled by the client, or a separate wallet-like nonce tracker thats not part of the app
func (app *BurrowMint) GetCheckCache() *sm.BlockCache {
	return app.checkCache
}

func NewBurrowMint(s *sm.State, evsw tendermint_events.EventSwitch,
	logger logging_types.InfoTraceLogger) *BurrowMint {
	return &BurrowMint{
		state:      s,
		cache:      sm.NewBlockCache(s),
		checkCache: sm.NewBlockCache(s),
		evc:        tendermint_events.NewEventCache(evsw),
		evsw:       evsw,
		logger:     logging.WithScope(logger, "BurrowMint"),
	}
}

// Implements manager/types.Application
func (app *BurrowMint) Info() (info abci.ResponseInfo) {
	return abci.ResponseInfo{}
}

// Implements manager/types.Application
func (app *BurrowMint) SetOption(key string, value string) (log string) {
	return ""
}

// Implements manager/types.Application
func (app *BurrowMint) DeliverTx(txBytes []byte) abci.Result {
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

	err = sm.ExecTx(app.cache, *tx, true, app.evc, app.logger)
	if err != nil {
		return abci.NewError(abci.CodeType_InternalError, fmt.Sprintf("Internal error: %v", err))
	}

	receipt := txs.GenerateReceipt(app.state.ChainID, *tx)
	receiptBytes := wire.BinaryBytes(receipt)
	return abci.NewResultOK(receiptBytes, "Success")
}

// Implements manager/types.Application
func (app *BurrowMint) CheckTx(txBytes []byte) abci.Result {
	var n int
	var err error
	tx := new(txs.Tx)
	buf := bytes.NewBuffer(txBytes)
	wire.ReadBinaryPtr(tx, buf, len(txBytes), &n, &err)
	if err != nil {
		return abci.NewError(abci.CodeType_EncodingError, fmt.Sprintf("Encoding error: %v", err))
	}

	// TODO: map ExecTx errors to sensible abci error codes
	err = sm.ExecTx(app.checkCache, *tx, false, nil, app.logger)
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
func (app *BurrowMint) Commit() (res abci.Result) {
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

	// TODO: [ben] over the tendermint 0.6 TMSP interface we have
	// no access to the block header implemented;
	// On Tendermint v0.8 load the blockheader into the application
	// state and remove the fixed 2-"seconds" per block internal clock.
	// NOTE: set internal time as two seconds per block
	app.state.LastBlockTime = app.state.LastBlockTime.Add(time.Duration(2) * time.Second)
	appHash := app.state.Hash()
	return abci.NewResultOK(appHash, "Success")
}

func (app *BurrowMint) Query(query []byte) (res abci.Result) {
	return abci.NewResultOK(nil, "Success")
}
