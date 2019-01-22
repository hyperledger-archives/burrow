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

	if err != nil {
		return err
	}

	err = state.IterateAccounts(func(acc *acm.Account) error {
		err = stream.Send(&dump.Dump{Height: height, Account: acc})
		if err != nil {
			return err
		}

		err = ds.state.IterateStorage(acc.Address, func(key, value binary.Word256) error {
			return stream.Send(&dump.Dump{
				Height: height,
				AccountStorage: &dump.AccountStorage{
					Address: acc.Address,
					Storage: &dump.Storage{Key: key, Value: value},
				},
			})
		})

		return nil
	})

	if err != nil {
		return err
	}

	err = state.IterateNames(func(entry *names.Entry) error {
		return stream.Send(&dump.Dump{Height: height, Name: entry})
	})

	if err != nil {
		return err
	}

	err = ds.state.IterateTx(0, height, func(tx *exec.TxExecution) error {
		for i := 0; i < len(tx.Events); i++ {
			event := tx.Events[i]
			if event.Log != nil {
				err := stream.Send(&dump.Dump{Height: event.Header.Height, EVMEvent: event.Log})
				if err != nil {
					return err
				}
			}
		}
		return nil
	})

	return nil
}
