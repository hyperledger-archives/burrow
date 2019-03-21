package rpcdump

import (
	"time"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/binary"
	dump "github.com/hyperledger/burrow/dump"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/execution/state"
	"github.com/hyperledger/burrow/logging"
)

type dumpServer struct {
	state      *state.State
	blockchain bcm.BlockchainInfo
	logger     *logging.Logger
}

var _ DumpServer = &dumpServer{}

func NewDumpServer(state *state.State, blockchain bcm.BlockchainInfo, logger *logging.Logger) *dumpServer {
	return &dumpServer{
		state:      state,
		blockchain: blockchain,
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

	err = st.IterateAccounts(func(acc *acm.Account) error {
		err = stream.Send(&dump.Dump{Height: height, Account: acc})
		if err != nil {
			return err
		}

		storage := dump.AccountStorage{
			Address: acc.Address,
			Storage: make([]*dump.Storage, 0),
		}

		err = st.IterateStorage(acc.Address, func(key, value binary.Word256) error {
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
	var origin *exec.Origin

	return ds.state.IterateStreamEvents(nil, &exec.StreamKey{Height: height},
		func(ev *exec.StreamEvent) error {
			switch {
			case ev.BeginBlock != nil:
				blockTime = ev.BeginBlock.Header.GetTime()
			case ev.BeginTx != nil:
				origin = ev.BeginTx.TxHeader.Origin
			case ev.Event != nil && ev.Event.Log != nil:
				evmevent := dump.EVMEvent{Event: ev.Event.Log}
				if origin != nil {
					// this event was already restored
					evmevent.ChainID = origin.ChainID
					evmevent.Time = origin.Time
				} else {
					// this event was generated on this chain
					evmevent.ChainID = ds.blockchain.ChainID()
					evmevent.Time = blockTime
				}
				err := stream.Send(&dump.Dump{Height: ev.Event.Header.Height, EVMEvent: &evmevent})
				if err != nil {
					return err
				}
			case ev.EndTx != nil:
				origin = nil
			}
			return nil
		})
}
