package tmsp

import (
	"bytes"
	"fmt"
	"sync"

	sm "github.com/eris-ltd/eris-db/state"
	types "github.com/eris-ltd/eris-db/txs"

	"github.com/tendermint/go-events"
	"github.com/tendermint/go-wire"

	tmsp "github.com/tendermint/tmsp/types"
)

//--------------------------------------------------------------------------------
// ErisDBApp holds the current state, runs transactions, computes hashes.
// Typically two connections are opened by the tendermint core:
// one for mempool, one for consensus.

type ErisDBApp struct {
	mtx sync.Mutex

	state *sm.State
	cache *sm.BlockCache

	evc  *events.EventCache
	evsw *events.EventSwitch
}

func (app *ErisDBApp) GetState() *sm.State {
	app.mtx.Lock()
	defer app.mtx.Unlock()
	return app.state.Copy()
}

func NewErisDBApp(s *sm.State, evsw *events.EventSwitch) *ErisDBApp {
	return &ErisDBApp{
		state: s,
		cache: sm.NewBlockCache(s),
		evc:   events.NewEventCache(evsw),
		evsw:  evsw,
	}
}

// Implements tmsp.Application
func (appC *ErisDBApp) Info() (info string) {
	return "ErisDB"
}

// Implements tmsp.Application
func (appC *ErisDBApp) SetOption(key string, value string) (log string) {
	return ""
}

// Implements tmsp.Application
func (appC ErisDBApp) AppendTx(txBytes []byte) (code tmsp.CodeType, result []byte, log string) {
	var n int
	var err error
	tx := new(types.Tx)
	buf := bytes.NewBuffer(txBytes)
	wire.ReadBinaryPtr(tx, buf, len(txBytes), &n, &err)
	if err != nil {
		return tmsp.CodeType_EncodingError, nil, fmt.Sprintf("Encoding error: %v", err)
	}

	err = sm.ExecTx(appC.cache, *tx, true, appC.evc)
	if err != nil {
		return tmsp.CodeType_InternalError, nil, fmt.Sprintf("Encoding error: %v", err)
	}

	return tmsp.CodeType_OK, nil, ""
}

// Implements tmsp.Application
func (appC ErisDBApp) CheckTx(txBytes []byte) (code tmsp.CodeType, result []byte, log string) {
	// TODO: precheck
	return tmsp.CodeType_OK, nil, ""
}

// Implements tmsp.Application
// GetHash should commit the state (called at end of block)
func (appC *ErisDBApp) GetHash() (hash []byte, log string) {
	appC.cache.Sync()

	// save state to disk
	appC.state.Save()

	// flush events to listeners (XXX: note issue with blocking)
	appC.evc.Flush()

	return appC.state.Hash(), ""
}

func (appC *ErisDBApp) Query(query []byte) (code tmsp.CodeType, result []byte, log string) {
	return tmsp.CodeType_OK, nil, ""
}
