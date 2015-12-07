package tmsp

import (
	"bytes"
	"sync"

	//sm "github.com/eris-ltd/eris-db/state"
	// txs "github.com/eris-ltd/eris-db/txs"

	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/events"
	sm "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/state"
	txs "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/types"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/wire"

	"github.com/tendermint/tmsp/types"
)

//--------------------------------------------------------------------------------
// ErisDBApp holds the current state and opens new contexts for proposed updates to the state

type ErisDBApp struct {
	mtx sync.Mutex

	state *sm.State
	evsw  *events.EventSwitch
}

func (app *ErisDBApp) GetState() *sm.State {
	app.mtx.Lock()
	defer app.mtx.Unlock()
	return app.state.Copy()
}

func NewErisDBApp(s *sm.State, evsw *events.EventSwitch) *ErisDBApp {
	return &ErisDBApp{
		state: s,
		evsw:  evsw,
	}
}

// Implements tmsp.Application
func (app *ErisDBApp) Open() types.AppContext {
	app.mtx.Lock()
	state := app.state.Copy()
	app.mtx.Unlock()
	return &ErisDBAppContext{
		state: state,
		cache: sm.NewBlockCache(state),
		evc:   events.NewEventCache(app.evsw),
	}
}

//--------------------------------------------------------------------------------
// ErisDBAppContext runs transactions, computes hashes, and handles commits/rollbacks.
// Typically two contexts (connections) are opened by the tendermint core:
// one for mempool, one for consensus.

type ErisDBAppContext struct {
	app *ErisDBApp

	state *sm.State
	cache *sm.BlockCache

	evc *events.EventCache
}

func (appC *ErisDBAppContext) Echo(message string) string {
	return message
}

func (appC *ErisDBAppContext) Info() []string {
	return []string{"ErisDB"}
}

func (appC *ErisDBAppContext) SetOption(key string, value string) types.RetCode {
	return 0
}

func (appC ErisDBAppContext) AppendTx(txBytes []byte) ([]types.Event, types.RetCode) {
	var n int64
	var err error
	tx := new(txs.Tx)
	buf := bytes.NewBuffer(txBytes)
	wire.ReadBinaryPtr(tx, buf, &n, &err)
	if err != nil {
		// TODO: handle error
		return nil, types.RetCodeEncodingError
	}

	err = sm.ExecTx(appC.cache, *tx, true, appC.evc)
	if err != nil {
		// TODO: handle error
		return nil, types.RetCodeInternalError // ?!
	}

	return nil, types.RetCodeOK
}

func (appC *ErisDBAppContext) GetHash() ([]byte, types.RetCode) {
	appC.cache.Sync()
	return appC.state.Hash(), types.RetCodeOK
}

func (appC *ErisDBAppContext) Commit() types.RetCode {
	// save state to disk
	appC.state.Save()

	// flush events to listeners
	appC.evc.Flush()

	// update underlying app state for new tmsp connections
	appC.app.mtx.Lock()
	appC.app.state = appC.state
	appC.app.mtx.Unlock()
	return 0
}

func (appC *ErisDBAppContext) Rollback() types.RetCode {
	appC.app.mtx.Lock()
	appC.state = appC.app.state
	appC.app.mtx.Unlock()

	appC.cache = sm.NewBlockCache(appC.state)
	return 0
}

func (appC *ErisDBAppContext) AddListener(key string) types.RetCode {
	return 0
}

func (appC *ErisDBAppContext) RemListener(key string) types.RetCode {
	return 0
}

func (appC *ErisDBAppContext) Close() error {
	return nil
}
