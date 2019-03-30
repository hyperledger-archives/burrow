package rpctest

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/rpc/rpcevents"
)

const maximumDurationWithoutProgress = time.Second

type TxExpecter struct {
	sync.Mutex
	emitter       *event.Emitter
	subID         string
	name          string
	all           map[string]struct{}
	expected      map[string]struct{}
	received      map[string]struct{}
	asserted      bool
	succeeded     chan struct{}
	ready         chan struct{}
	previousTotal int
	blockRange    *rpcevents.BlockRange
}

// Start listening for blocks and cross off any transactions that were expected.
// Expect can be called multiple times before a single call to AssertCommitted.
// TxExpecter is single-shot - create multiple TxExpecters if you want to call AssertCommitted multiple times.
func ExpectTxs(emitter *event.Emitter, name string) *TxExpecter {
	exp := &TxExpecter{
		emitter:    emitter,
		subID:      event.GenSubID(),
		name:       name,
		all:        make(map[string]struct{}),
		expected:   make(map[string]struct{}),
		received:   make(map[string]struct{}),
		succeeded:  make(chan struct{}),
		ready:      make(chan struct{}),
		blockRange: &rpcevents.BlockRange{},
	}
	go exp.listen()
	<-exp.ready
	return exp
}

// Expect a transaction to be committed
func (exp *TxExpecter) Expect(txHash binary.HexBytes) {
	exp.Lock()
	defer exp.Unlock()
	if exp.closed() {
		panic(fmt.Errorf("cannot call Expect after AssertCommitted"))
	}
	key := txHash.String()
	exp.expected[key] = struct{}{}
	exp.all[key] = struct{}{}
}

// Assert that all expected transactions are committed. Will block until all expected transactions are committed.
// Returns the BlockRange over which the transactions were committed.
func (exp *TxExpecter) AssertCommitted(t testing.TB) *rpcevents.BlockRange {
	exp.Lock()
	// close() clears subID to indicate this TxExpecter ha been used
	if exp.closed() {
		panic(fmt.Errorf("cannot call AssertCommitted more than once"))
	}
	exp.asserted = true
	if exp.reconcile() {
		return exp.blockRange
	}
	exp.Unlock()
	defer exp.close()
	var err error
	for err == nil {
		select {
		case <-exp.succeeded:
			return exp.blockRange
		case <-time.After(maximumDurationWithoutProgress):
			err = exp.assertMakingProgress()
		}
	}
	t.Fatal(err)
	return nil
}

func (exp *TxExpecter) listen() {
	numTxs := 0
	ch, err := exp.emitter.Subscribe(context.Background(), exp.subID, exec.QueryForBlockExecution(), 1)
	if err != nil {
		panic(fmt.Errorf("ExpectTxs(): could not subscribe to blocks: %v", err))
	}
	close(exp.ready)
	defer exp.close()
	for msg := range ch {
		be := msg.(*exec.BlockExecution)
		blockTxs := len(be.TxExecutions)
		numTxs += blockTxs
		fmt.Printf("%s: Total TXs committed at block %v: %v (+%v)\n", exp.name, be.GetHeight(), numTxs, blockTxs)
		for _, txe := range be.TxExecutions {
			// Return if this is the last expected transaction (and we are finished expecting)
			if exp.receive(txe) {
				return
			}
		}
	}
}

func (exp *TxExpecter) close() {
	exp.Lock()
	defer exp.Unlock()
	if !exp.closed() {
		close(exp.succeeded)
		exp.emitter.UnsubscribeAll(context.Background(), exp.subID)
		exp.subID = ""
	}
}

func (exp *TxExpecter) closed() bool {
	return exp.subID == ""
}

func (exp *TxExpecter) receive(txe *exec.TxExecution) (done bool) {
	exp.Lock()
	defer exp.Unlock()
	exp.received[txe.TxHash.String()] = struct{}{}
	if exp.blockRange.Start == nil {
		exp.blockRange.Start = rpcevents.AbsoluteBound(txe.Height)
		exp.blockRange.End = rpcevents.AbsoluteBound(txe.Height)
	}
	exp.blockRange.End.Index = txe.Height
	if exp.asserted {
		return exp.reconcile()
	}
	return false
}

func (exp *TxExpecter) reconcile() (done bool) {
	for re := range exp.received {
		if _, ok := exp.expected[re]; ok {
			// Remove from expected
			delete(exp.expected, re)
			// No longer need to cache in received
			delete(exp.received, re)
		}
	}
	total := len(exp.expected)
	return total == 0
}

func (exp *TxExpecter) assertMakingProgress() error {
	exp.Lock()
	defer exp.Unlock()
	total := len(exp.expected)
	if exp.previousTotal == 0 {
		exp.previousTotal = total
		return nil
	}
	// if the total is reducing we are making progress
	if total < exp.previousTotal {
		return nil
	}
	committed := total - len(exp.all)
	committedString := "none"
	if committed != 0 {
		committedString = fmt.Sprintf("only %d", committed)
	}
	return fmt.Errorf("TxExpecter timed out after %v: expecting %d txs to be committed but %s were",
		maximumDurationWithoutProgress, total, committedString)
}
