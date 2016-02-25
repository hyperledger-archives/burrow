package mempool

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tendermint/go-clist"

	sm "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/state"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/types"
)

/*

The mempool pushes new txs onto the proxyAppConn.
It gets a stream of (req, res) tuples from the proxy.
The memool stores good txs in a concurrent linked-list.

Multiple concurrent go-routines can traverse this linked-list
safely by calling .NextWait() on each element.

So we have several go-routines:
1. Consensus calling Update() and Reap() synchronously
2. Many mempool reactor's peer routines calling CheckTx()
3. Many mempool reactor's peer routines traversing the txs linked list
4. Another goroutine calling GarbageCollectTxs() periodically

To manage these goroutines, there are three methods of locking.
1. Mutations to the linked-list is protected by an internal mtx (CList is goroutine-safe)
2. Mutations to the linked-list elements are atomic
3. CheckTx() calls can be paused upon Update() and Reap(), protected by .mtx

Garbage collection of old elements from mempool.txs is handlde via
the DetachPrev() call, which makes old elements not reachable by
peer broadcastTxRoutine() automatically garbage collected.

TODO: Better handle tmsp client errors. (make it automatically handle connection errors)

*/

const cacheSize = 100000

type Mempool struct {
	mtx   sync.Mutex
	state *sm.State
	cache *sm.BlockCache

	txs           *clist.CList    // concurrent linked-list of good txs
	counter       int64           // simple incrementing counter
	height        int             // the last block Update()'d to
	rechecking    int32           // for re-checking filtered txs on Update()
	recheckCursor *clist.CElement // next expected response
	recheckEnd    *clist.CElement // re-checking stops here

	// Keep a cache of already-seen txs.
	// This reduces the pressure on the proxyApp.
	cacheMap  map[string]struct{}
	cacheList *list.List
}

func NewMempool(state *sm.State) *Mempool {
	mempool := &Mempool{
		state:         state,
		cache:         sm.NewBlockCache(state),
		txs:           clist.New(),
		counter:       0,
		height:        0,
		rechecking:    0,
		recheckCursor: nil,
		recheckEnd:    nil,

		cacheMap:  make(map[string]struct{}, cacheSize),
		cacheList: list.New(),
	}
	return mempool
}

// Return the first element of mem.txs for peer goroutines to call .NextWait() on.
// Blocks until txs has elements.
func (mem *Mempool) TxsFrontWait() *clist.CElement {
	return mem.txs.FrontWait()
}

func (mem *Mempool) TxID(tx types.Tx) string {
	return string(types.TxID(mem.state.ChainID, tx))
}

// Try a new transaction in the mempool.
// Potentially blocking if we're blocking on Update() or Reap().
func (mem *Mempool) AddTx(tx types.Tx) (err error) {
	mem.mtx.Lock()
	defer mem.mtx.Unlock()

	// CACHE
	if _, exists := mem.cacheMap[mem.TxID(tx)]; exists {
		return nil
	}
	if mem.cacheList.Len() >= cacheSize {
		popped := mem.cacheList.Front()
		poppedTx := popped.Value.(types.Tx)
		delete(mem.cacheMap, mem.TxID(poppedTx))
		mem.cacheList.Remove(popped)
	}
	mem.cacheMap[mem.TxID(tx)] = struct{}{}
	mem.cacheList.PushBack(tx)
	// END CACHE

	err = sm.ExecTx(mem.cache, tx, false, nil)
	if err != nil {
		log.Info("AddTx() error", "tx", tx, "error", err)
		return err
	} else {
		log.Info("AddTx() success", "tx", tx)
		mem.counter++
		memTx := &mempoolTx{
			counter: mem.counter,
			height:  int64(mem.height),
			tx:      tx,
		}
		mem.txs.PushBack(memTx)
		return nil
	}
	return nil
}

func (mem *Mempool) GetState() *sm.State {
	return mem.state
}

func (mem *Mempool) GetCache() *sm.BlockCache {
	return mem.cache
}

func (mem *Mempool) GetHeight() int {
	mem.mtx.Lock()
	defer mem.mtx.Unlock()
	return mem.state.LastBlockHeight
}

// Get the valid transactions remaining
func (mem *Mempool) GetProposalTxs() []types.Tx {
	mem.mtx.Lock()
	defer mem.mtx.Unlock()

	for atomic.LoadInt32(&mem.rechecking) > 0 {
		// TODO: Something better?
		time.Sleep(time.Millisecond * 10)
	}

	txs := mem.collectTxs()
	return txs
}

func (mem *Mempool) collectTxs() []types.Tx {
	txs := make([]types.Tx, 0, mem.txs.Len())
	for e := mem.txs.Front(); e != nil; e = e.Next() {
		memTx := e.Value.(*mempoolTx)
		txs = append(txs, memTx.tx)
	}
	return txs
}

// Tell mempool that these txs were committed.
// Mempool will discard these txs.
// NOTE: this should be called *after* block is committed by consensus.
func (mem *Mempool) ResetForBlockAndState(block *types.Block, state *sm.State) {
	mem.mtx.Lock()
	defer mem.mtx.Unlock()

	mem.state = state.Copy()
	mem.cache = sm.NewBlockCache(mem.state)

	// First, create a lookup map of txns in new txs.
	txsMap := make(map[string]struct{})
	for _, tx := range block.Data.Txs {
		txsMap[mem.TxID(tx)] = struct{}{}
	}

	// Set height
	mem.height = block.Height
	// Remove transactions that are already in txs.
	goodTxs := mem.filterTxs(txsMap)
	// Recheck mempool txs
	// TODO: make optional
	mem.recheckTxs(goodTxs)

	// At this point, mem.txs are being rechecked.
	// mem.recheckCursor re-scans mem.txs and possibly removes some txs.
	// Before mem.Reap(), we should wait for mem.recheckCursor to be nil.
}

func (mem *Mempool) filterTxs(blockTxsMap map[string]struct{}) []types.Tx {
	goodTxs := make([]types.Tx, 0, mem.txs.Len())
	for e := mem.txs.Front(); e != nil; e = e.Next() {
		memTx := e.Value.(*mempoolTx)
		if _, ok := blockTxsMap[mem.TxID(memTx.tx)]; ok {
			// Remove the tx since already in block.
			mem.txs.Remove(e)
			e.DetachPrev()
			continue
		}
		// Good tx!
		goodTxs = append(goodTxs, memTx.tx)
	}
	return goodTxs
}

// NOTE: pass in goodTxs because mem.txs can mutate concurrently.
func (mem *Mempool) recheckTxs(goodTxs []types.Tx) {
	if len(goodTxs) == 0 {
		return
	}
	atomic.StoreInt32(&mem.rechecking, 1)
	mem.recheckCursor = mem.txs.Front()
	mem.recheckEnd = mem.txs.Back()

	for _, tx := range goodTxs {
		err := sm.ExecTx(mem.cache, tx, false, nil)
		if err != nil {
			// Tx became invalidated due to newly committed block.
			mem.txs.Remove(mem.recheckCursor)
			mem.recheckCursor.DetachPrev()
		}
		if mem.recheckCursor == mem.recheckEnd {
			mem.recheckCursor = nil
		} else {
			mem.recheckCursor = mem.recheckCursor.Next()
		}
		if mem.recheckCursor == nil {
			// Done!
			atomic.StoreInt32(&mem.rechecking, 0)
		}
	}
}

//--------------------------------------------------------------------------------

// A transaction that successfully ran
type mempoolTx struct {
	counter int64    // a simple incrementing counter
	height  int64    // height that this tx had been validated in
	tx      types.Tx //
}

func (memTx *mempoolTx) Height() int {
	return int(atomic.LoadInt64(&memTx.height))
}
