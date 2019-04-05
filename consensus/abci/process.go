package abci

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"github.com/hyperledger/burrow/bcm"

	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/txs"
	"github.com/tendermint/tendermint/abci/types"
	tmTypes "github.com/tendermint/tendermint/types"
)

type Process struct {
	ticker       *time.Ticker
	committer    execution.BatchCommitter
	blockchain   *bcm.Blockchain
	done         chan struct{}
	panic        func(error)
	commitNeeded bool
	txDecoder    txs.Decoder
	shutdownOnce sync.Once
}

// NewProcess returns a no-consensus ABCI process suitable for running a single node without Tendermint.
// The CheckTx function can be used to submit transactions which are processed according
func NewProcess(committer execution.BatchCommitter, blockchain *bcm.Blockchain, txDecoder txs.Decoder,
	commitInterval time.Duration, panicFunc func(error)) *Process {

	p := &Process{
		committer:  committer,
		blockchain: blockchain,
		done:       make(chan struct{}),
		txDecoder:  txDecoder,
		panic:      panicFunc,
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
	// FIXME: [Silas] this means that any transaction that a transaction that fails CheckTx
	// that would not normally end up stored in state (as an exceptional tx) will get stored in state.
	// This means that the same sequence of transactions fed to no consensus mode can give rise to a state with additional
	// invalid transactions in state. Since the state hash is non-deterministic based on when the commits happen it's not
	// clear this is a problem. The underlying state will be compatible.
	checkTx := ExecuteTx(header, p.committer, p.txDecoder, tx)
	cb(types.ToResponseCheckTx(checkTx))
	p.commitNeeded = true
	if p.ticker == nil {
		err := p.commit()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Process) Shutdown(ctx context.Context) (err error) {
	p.committer.Lock()
	defer p.committer.Unlock()
	p.shutdownOnce.Do(func() {
		if p.ticker != nil {
			p.ticker.Stop()
		}
		close(p.done)
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
	p.committer.Lock()
	defer p.committer.Unlock()
	err := p.commit()
	if err != nil {
		p.panic(err)
	}
}

func (p *Process) commit() error {
	const errHeader = "commit():"
	if !p.commitNeeded {
		return nil
	}

	appHash, err := p.committer.Commit(nil)
	if err != nil {
		return fmt.Errorf("%s could not Commit tx %v", errHeader, err)
	}

	// Maintain a basic hashed linked list, mixing in the appHash as we go
	hasher := sha256.New()
	hasher.Write(appHash)
	hasher.Write(p.blockchain.LastBlockHash())

	err = p.blockchain.CommitBlock(time.Now(), hasher.Sum(nil), appHash)
	if err != nil {
		return fmt.Errorf("%s could not CommitBlock %v", errHeader, err)
	}
	p.commitNeeded = false
	return nil
}
