package abci

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/txs"
	"github.com/tendermint/tendermint/abci/types"
	tmTypes "github.com/tendermint/tendermint/types"
)

const txBufferSize = 100

type Process struct {
	ticker       *time.Ticker
	committer    execution.BatchCommitter
	txs          chan tmTypes.Tx
	done         chan struct{}
	panic        func(error)
	commitNeeded bool
	txDecoder    txs.Decoder
	shutdownOnce sync.Once
}

// NewProcess returns a no-consensus ABCI process suitable for running a single node without Tendermint.
// The CheckTx function can be used to submit transactions which are processed according
func NewProcess(committer execution.BatchCommitter, txDecoder txs.Decoder, commitInterval time.Duration,
	panicFunc func(error)) *Process {

	p := &Process{
		committer: committer,
		txs:       make(chan tmTypes.Tx),
		done:      make(chan struct{}),
		txDecoder: txDecoder,
		panic:     panicFunc,
	}

	if commitInterval != 0 {
		p.ticker = time.NewTicker(commitInterval)
		go p.triggerCommits()
	}

	return p
}

func (p *Process) CheckTx(tx tmTypes.Tx, cb func(*types.Response)) error {
	const header = "DeliverTx"
	p.committer.Lock()
	defer p.committer.Unlock()
	// Skip check - deliver immediately
	checkTx := ExecuteTx(header, p.committer, p.txDecoder, tx)
	cb(types.ToResponseCheckTx(checkTx))
	p.commitNeeded = true
	return nil
}

func (p *Process) Shutdown(ctx context.Context) (err error) {
	p.committer.Lock()
	defer p.committer.Unlock()
	p.shutdownOnce.Do(func() {
		p.ticker.Stop()
		close(p.txs)
		select {
		case <-p.done:
		case <-ctx.Done():
			err = ctx.Err()
		}
	})
	return
}

func (p *Process) triggerCommits() {
	for {
		select {
		case <-p.ticker.C:
			p.commitOrPanic()
		case <-p.done:
			// Escape loop since ticket channel is never closed
		}
	}
}

func (p *Process) commitOrPanic() {
	err := p.commit()
	if err != nil {
		p.panic(err)
	}
}

func (p *Process) commit() error {
	const errHeader = "commit():"
	p.committer.Lock()
	defer p.committer.Unlock()
	if !p.commitNeeded {
		return nil
	}

	_, err := p.committer.Commit(nil)
	if err != nil {
		return fmt.Errorf("%s could not Commit tx %v", errHeader, err)
	}

	p.commitNeeded = false
	return nil
}
