package execution

import (
	"testing"
	"time"

	"github.com/hyperledger/burrow/account/state"
	"github.com/hyperledger/burrow/blockchain"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/txs"
	"github.com/tendermint/abci/types"
)

func TestTransactor_TransactAndHold(t *testing.T) {
}

type testTransactor struct {
	ResponseCh chan<- *types.Response
	state.IterableWriter
	event.Emitter
	*Transactor
}

func newTestTransactor(txProcessor func(tx txs.Tx) (*types.Response, error)) testTransactor {
	st := state.NewMemoryState()
	emitter := event.NewEmitter(logger)
	trans := NewTransactor(blockchain.NewTip(testChainID, time.Time{}, nil),
		emitter, func(tx txs.Tx, callback func(res *types.Response)) error {
			res, err := txProcessor(tx)
			if err != nil {
				return err
			}
			callback(res)
			return nil
		}, logger)

	return testTransactor{
		IterableWriter: st,
		Emitter:        emitter,
		Transactor:     trans,
	}
}

func TestTransactor_Transact(t *testing.T) {
	//trans := newTestTransactor()
}
