package tmsp

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sync"

	sm "github.com/eris-ltd/eris-db/state"
	types "github.com/eris-ltd/eris-db/txs"

	"github.com/tendermint/go-events"
	client "github.com/tendermint/go-rpc/client"
	"github.com/tendermint/go-wire"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	tmsp "github.com/tendermint/tmsp/types"
)

//--------------------------------------------------------------------------------
// ErisDBApp holds the current state, runs transactions, computes hashes.
// Typically two connections are opened by the tendermint core:
// one for mempool, one for consensus.

type ErisDBApp struct {
	mtx sync.Mutex

	state      *sm.State
	cache      *sm.BlockCache
	checkCache *sm.BlockCache // for CheckTx (eg. so we get nonces right)

	evc  *events.EventCache
	evsw *events.EventSwitch

	// client to the tendermint core rpc
	client *client.ClientURI
	host   string // tendermint core endpoint

	nTxs int // count txs in a block
}

func (app *ErisDBApp) GetState() *sm.State {
	app.mtx.Lock()
	defer app.mtx.Unlock()
	return app.state.Copy()
}

// TODO: this is used for call/callcode and to get nonces during mempool.
// the former should work on last committed state only and the later should
// be handled by the client, or a separate wallet-like nonce tracker thats not part of the app
func (app *ErisDBApp) GetCheckCache() *sm.BlockCache {
	return app.checkCache
}

func (app *ErisDBApp) SetHostAddress(host string) {
	app.host = host
	app.client = client.NewClientURI(host) //fmt.Sprintf("http://%s", host))
}

// Broadcast a tx to the tendermint core
// NOTE: this assumes we know the address of core
func (app *ErisDBApp) BroadcastTx(tx types.Tx) error {
	buf := new(bytes.Buffer)
	var n int
	var err error
	wire.WriteBinary(struct{ types.Tx }{tx}, buf, &n, &err)
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

func NewErisDBApp(s *sm.State, evsw *events.EventSwitch) *ErisDBApp {
	return &ErisDBApp{
		state:      s,
		cache:      sm.NewBlockCache(s),
		checkCache: sm.NewBlockCache(s),
		evc:        events.NewEventCache(evsw),
		evsw:       evsw,
	}
}

// Implements tmsp.Application
func (app *ErisDBApp) Info() (info string) {
	return "ErisDB"
}

// Implements tmsp.Application
func (app *ErisDBApp) SetOption(key string, value string) (log string) {
	return ""
}

// Implements tmsp.Application
func (app *ErisDBApp) AppendTx(txBytes []byte) (res tmsp.Result) {

	app.nTxs += 1

	// XXX: if we had tx ids we could cache the decoded txs on CheckTx
	var n int
	var err error
	tx := new(types.Tx)
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
	// TODO: need to return receipt so rpc.ResultBroadcastTx.Data (or Log) is the receipt
	return tmsp.NewResultOK(nil, "Success")
}

// Implements tmsp.Application
func (app *ErisDBApp) CheckTx(txBytes []byte) (res tmsp.Result) {
	var n int
	var err error
	tx := new(types.Tx)
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

	// TODO: need to return receipt so rpc.ResultBroadcastTx.Data (or Log) is the receipt
	return tmsp.NewResultOK(nil, "Success")
}

// Implements tmsp.Application
// Commit the state (called at end of block)
// NOTE: CheckTx/AppendTx must not run concurrently with Commit -
//	the mempool should run during AppendTxs, but lock for Commit and Update
func (app *ErisDBApp) Commit() (res tmsp.Result) {
	app.mtx.Lock() // the lock protects app.state
	defer app.mtx.Unlock()

	app.state.LastBlockHeight += 1
	log.Info("Commit", "block", app.state.LastBlockHeight)

	// sync the AppendTx cache
	app.cache.Sync()

	// if there were any txs in the block,
	// reset the check cache to the new height
	if app.nTxs > 0 {
		log.Info("Reset checkCache", "txs", app.nTxs)
		app.checkCache = sm.NewBlockCache(app.state)
	}
	app.nTxs = 0

	// save state to disk
	app.state.Save()

	// flush events to listeners (XXX: note issue with blocking)
	app.evc.Flush()

	return tmsp.NewResultOK(app.state.Hash(), "Success")
}

func (app *ErisDBApp) Query(query []byte) (res tmsp.Result) {
	return tmsp.NewResultOK(nil, "Success")
}
