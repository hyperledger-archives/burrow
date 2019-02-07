package rpcdump

import (
	"time"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/consensus/tendermint"
	dump "github.com/hyperledger/burrow/dump"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/execution/state"
	"github.com/hyperledger/burrow/logging"
)

type dumpServer struct {
	state      *state.State
	blockchain bcm.BlockchainInfo
	nodeView   *tendermint.NodeView
	logger     *logging.Logger
}

var _ DumpServer = &dumpServer{}

func NewDumpServer(state *state.State, blockchain bcm.BlockchainInfo, nodeView *tendermint.NodeView, logger *logging.Logger) *dumpServer {
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
	st, err := ds.state.LoadHeight(height)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	err = st.IterateAccounts(func(acc *acm.Account) error {
		err = stream.Send(&dump.Dump{Height: height, Account: acc})
		if err != nil {
			return err
		}

		storage := dump.AccountStorage{
			Address: acc.Address,
			Storage: make([]*dump.Storage, 0),
		}

		err = ds.state.IterateStorage(acc.Address, func(key, value binary.Word256) error {
			storage.Storage = append(storage.Storage, &dump.Storage{Key: key, Value: value})
			return nil
		})

		if err != nil {
			return err
		}

		if len(storage.Storage) > 0 {
			return stream.Send(&dump.Dump{
				Height:         height,
				AccountStorage: &storage,
			})
		}

		return nil
	})

	if err != nil {
		return err
	}

	err = st.IterateNames(func(entry *names.Entry) error {
		return stream.Send(&dump.Dump{Height: height, Name: entry})
	})

	if err != nil {
		return err
	}

	var blockTime time.Time

	return ds.state.IterateStreamEvents(0, height, func(ev *exec.StreamEvent) error {
		if ev.BeginBlock != nil {
			blockTime = ev.BeginBlock.Header.GetTime()
		}
		txe := ev.TxExecution
		if txe == nil {
			return nil
		}
		for _, event := range ev.TxExecution.Events {
			if event.Log != nil {
				evmevent := dump.EVMEvent{Event: event.Log}
				if txe.Origin != nil {
					// this event was already restored
					evmevent.ChainID = txe.Origin.ChainID
					evmevent.Time = txe.Origin.Time
				} else {
					// this event was generated on this chain
					evmevent.ChainID = ds.blockchain.ChainID()
					evmevent.Time = blockTime
				}
				err := stream.Send(&dump.Dump{Height: event.Header.Height, EVMEvent: &evmevent})
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}
