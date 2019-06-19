package dump

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/execution/state"
	"github.com/hyperledger/burrow/logging"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/db"
)

var cdc = amino.NewCodec()

type Option uint64

const (
	// Whether to send/receive these classes of data
	Accounts Option = 1 << iota
	Names
	Events
)

const (
	None Option = 0
	All         = Accounts | Names | Events
)

func (options Option) Enabled(option Option) bool {
	return options&option > 0
}

type Dumper struct {
	state      *state.State
	blockchain bcm.BlockchainInfo
	logger     *logging.Logger
}

func NewDumper(state *state.State, blockchain bcm.BlockchainInfo, logger *logging.Logger) *Dumper {
	return &Dumper{
		state:      state,
		blockchain: blockchain,
		logger:     logger,
	}
}

func (ds *Dumper) Transmit(stream Sender, startHeight, endHeight uint64, options Option) error {
	height := endHeight
	if height == 0 {
		height = ds.blockchain.LastBlockHeight()
	}
	st, err := ds.state.LoadHeight(height)
	if err != nil {
		return err
	}

	if options.Enabled(Accounts) {
		ds.logger.InfoMsg("Dumping accounts")
		err = st.IterateAccounts(func(acc *acm.Account) error {
			err = stream.Send(&Dump{Height: height, Account: acc})
			if err != nil {
				return err
			}

			storage := AccountStorage{
				Address: acc.Address,
				Storage: make([]*Storage, 0),
			}

			err = st.IterateStorage(acc.Address, func(key, value binary.Word256) error {
				storage.Storage = append(storage.Storage, &Storage{Key: key, Value: value})
				return nil
			})

			if err != nil {
				return err
			}

			if len(storage.Storage) > 0 {
				return stream.Send(&Dump{
					Height:         height,
					AccountStorage: &storage,
				})
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	if options.Enabled(Names) {
		ds.logger.InfoMsg("Dumping names")
		err = st.IterateNames(func(entry *names.Entry) error {
			return stream.Send(&Dump{Height: height, Name: entry})
		})
		if err != nil {
			return err
		}
	}

	if options.Enabled(Events) {
		ds.logger.InfoMsg("Dumping events")
		var blockTime time.Time
		var origin *exec.Origin

		// Only return events from specified start height - allows for resume
		err = ds.state.IterateStreamEvents(&startHeight, &endHeight,
			func(ev *exec.StreamEvent) error {
				switch {
				case ev.BeginBlock != nil:
					ds.logger.TraceMsg("BeginBlock", "height", ev.BeginBlock.Height)
					blockTime = ev.BeginBlock.Header.GetTime()
				case ev.BeginTx != nil:
					origin = ev.BeginTx.TxHeader.Origin
				case ev.Event != nil && ev.Event.Log != nil:
					evmevent := EVMEvent{Event: ev.Event.Log}
					if origin != nil {
						// this event was already restored
						evmevent.ChainID = origin.ChainID
						evmevent.Time = origin.Time
					} else {
						// this event was generated on this chain
						evmevent.ChainID = ds.blockchain.ChainID()
						evmevent.Time = blockTime
					}
					err := stream.Send(&Dump{Height: ev.Event.Header.Height, EVMEvent: &evmevent})
					if err != nil {
						return err
					}
				case ev.EndTx != nil:
					origin = nil
				}
				return nil
			})
		if err != nil {
			return err
		}
	}

	return nil
}

func (ds *Dumper) Pipe(startHeight, endHeight uint64, options Option) Pipe {
	p := make(Pipe)
	go func() {
		err := ds.Transmit(p, startHeight, endHeight, options)
		if err != nil {
			p <- msg{err: err}
		}
		close(p)
	}()
	return p
}

func Write(stream Receiver, out io.Writer, useJSON bool, options Option) error {
	st := state.NewState(db.NewMemDB())
	_, _, err := st.Update(func(ws state.Updatable) error {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("failed to recv dump: %v", err)
			}

			if options.Enabled(Accounts) {
				// update our temporary state
				if resp.Account != nil {
					err := ws.UpdateAccount(resp.Account)
					if err != nil {
						return err
					}
				}
				if resp.AccountStorage != nil {
					for _, storage := range resp.AccountStorage.Storage {
						err := ws.SetStorage(resp.AccountStorage.Address, storage.Key, storage.Value)
						if err != nil {
							return err
						}
					}
				}
			}

			if options.Enabled(Names) {
				if resp.Name != nil {
					err := ws.UpdateName(resp.Name)
					if err != nil {
						return err
					}
				}
			}

			var bs []byte
			if useJSON {
				bs, err = json.Marshal(resp)
				if bs != nil {
					bs = append(bs, []byte("\n")...)
				}
			} else {
				bs, err = cdc.MarshalBinaryLengthPrefixed(resp)
			}
			if err != nil {
				return fmt.Errorf("failed to marshall dump: %v", err)
			}

			n, err := out.Write(bs)
			if err == nil && n < len(bs) {
				return fmt.Errorf("failed to write dump: %v", err)
			}
		}

		return nil
	})
	return err
}
