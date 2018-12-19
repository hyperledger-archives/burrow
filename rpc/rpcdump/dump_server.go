package rpcdump

import (
	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/consensus/tendermint"
	dump "github.com/hyperledger/burrow/dump"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/logging"
)

type dumpServer struct {
	state      *execution.State
	blockchain bcm.BlockchainInfo
	nodeView   *tendermint.NodeView
	logger     *logging.Logger
}

var _ DumpServer = &dumpServer{}

func NewDumpServer(state *execution.State, blockchain bcm.BlockchainInfo, nodeView *tendermint.NodeView, logger *logging.Logger) *dumpServer {
	return &dumpServer{
		state:      state,
		blockchain: blockchain,
		nodeView:   nodeView,
		logger:     logger,
	}
}

func (ds *dumpServer) GetDump(param *GetDumpParam, stream Dump_GetDumpServer) error {
	height := param.Height
	if height <= 0 {
		height = ds.blockchain.LastBlockHeight()
	}
	state, err := ds.state.LoadHeight(height)
	if err != nil {
		return err
	}

	err = stream.Send(&dump.Dump{
		Height: &dump.Height{Height: height},
	})

	if err != nil {
		return err
	}

	_, err = state.IterateAccounts(func(acc *acm.Account) (stopped bool) {
		stream.Send(&dump.Dump{Account: acc})

		stopped, err = ds.state.IterateStorage(acc.Address, func(key, value binary.Word256) (stopped bool) {
			stream.Send(&dump.Dump{
				AccountStorage: &dump.AccountStorage{
					Address: acc.Address,
					Storage: &dump.Storage{Key: key, Value: value},
				},
			})
			return
		})

		if err != nil {
			stopped = true
		}

		return false
	})

	if err != nil {
		return err
	}

	_, err = state.IterateNames(func(entry *names.Entry) (stop bool) {
		stream.Send(&dump.Dump{Name: entry})
		return
	})

	if err != nil {
		return err
	}

	_, err = ds.state.IterateTx(0, height, func(tx *exec.TxExecution) (stop bool) {
		for i := 0; i < len(tx.Events); i++ {
			event := tx.Events[i].GetLog()
			if event != nil {
				stream.Send(&dump.Dump{EVMEvent: &dump.EVMEvent{Height: tx.Height, Event: event}})
			}
		}
		return
	})

	return nil
}
