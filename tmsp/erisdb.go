package tmsp

import (
	"bytes"
	"sync"

	//sm "github.com/eris-ltd/eris-db/state"
	// txs "github.com/eris-ltd/eris-db/txs"

	sm "github.com/eris-ltd/eris-db/state"
	types "github.com/eris-ltd/eris-db/txs"

	"github.com/tendermint/go-wire"
	"github.com/tendermint/tendermint/events"

	tmsp "github.com/tendermint/tmsp/types"
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
func (app *ErisDBApp) Open() tmsp.AppContext {
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

func (appC *ErisDBAppContext) SetOption(key string, value string) tmsp.RetCode {
	return 0
}

func (appC ErisDBAppContext) AppendTx(txBytes []byte) ([]tmsp.Event, tmsp.RetCode) {
	var n int
	var err error
	tx := new(types.Tx)
	buf := bytes.NewBuffer(txBytes)
	wire.ReadBinaryPtr(tx, buf, len(txBytes), &n, &err)
	if err != nil {
		// TODO: handle error
		return nil, tmsp.RetCodeEncodingError
	}

	err = sm.ExecTx(appC.cache, *tx, true, appC.evc)
	if err != nil {
		// TODO: handle error
		return nil, tmsp.RetCodeInternalError // ?!
	}

	return nil, tmsp.RetCodeOK
}

func (appC *ErisDBAppContext) GetHash() ([]byte, tmsp.RetCode) {
	appC.cache.Sync()
	return appC.state.Hash(), tmsp.RetCodeOK
}

func (appC *ErisDBAppContext) Commit() tmsp.RetCode {
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

func (appC *ErisDBAppContext) Rollback() tmsp.RetCode {
	appC.app.mtx.Lock()
	appC.state = appC.app.state
	appC.app.mtx.Unlock()

	appC.cache = sm.NewBlockCache(appC.state)
	return 0
}

func (appC *ErisDBAppContext) AddListener(key string) tmsp.RetCode {
	return 0
}

func (appC *ErisDBAppContext) RemListener(key string) tmsp.RetCode {
	return 0
}

func (appC *ErisDBAppContext) Close() error {
	return nil
}
