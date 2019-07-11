package dump

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/encoding"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/execution/state"
	"github.com/hyperledger/burrow/logging"
	"github.com/tendermint/tendermint/libs/db"
)

const (
	// Whether to send/receive these classes of data
	Accounts Option = 1 << iota
	Names
	Events
)

// Chunk account storage into rows that are less than 1 MiB
const thresholdAccountStorageBytesPerRow = 1 << 20

type Sink interface {
	Send(*Dump) error
}

type Blockchain interface {
	ChainID() string
	LastBlockHeight() uint64
}

type Dumper struct {
	state      *state.State
	blockchain Blockchain
	logger     *logging.Logger
}

// Return a Dumper that can Transmit Dump rows to a Sink by pulling them out of the the provided State
func NewDumper(state *state.State, blockchain Blockchain) *Dumper {
	return &Dumper{
		state:      state,
		blockchain: blockchain,
		logger:     logging.NewNoopLogger(),
	}
}

type Option uint64

const (
	None Option = 0
	All         = Accounts | Names | Events
)

func (options Option) Enabled(option Option) bool {
	return options&option > 0
}

// Transmit Dump rows to the provided Sink over the inclusive range of heights provided, if endHeight is 0 the latest
// height is used.

func (ds *Dumper) Transmit(sink Sink, startHeight, endHeight uint64, options Option) error {
	lastHeight := ds.blockchain.LastBlockHeight()
	if endHeight == 0 || endHeight > lastHeight {
		endHeight = lastHeight
	}
	st, err := ds.state.LoadHeight(endHeight)
	if err != nil {
		return err
	}

	if options.Enabled(Accounts) {
		ds.logger.InfoMsg("Dumping accounts")
		err = st.IterateAccounts(func(acc *acm.Account) error {
			// Since we tend to want to handle accounts and their storage as a single unit we multiplex account
			// and storage within the same row. If the storage gets too large we chunk it and send in separate rows
			// (so that we stay well below the 4MiB GRPC message size limit and generally maintain stream-ability)
			row := &Dump{
				Height:  endHeight,
				Account: acc,
				AccountStorage: &AccountStorage{
					Address: acc.Address,
					Storage: make([]*Storage, 0),
				},
			}

			var storageBytes int
			err = st.IterateStorage(acc.Address, func(key binary.Word256, value []byte) error {
				if storageBytes > thresholdAccountStorageBytesPerRow {
					// Send the current row
					err = sink.Send(row)
					if err != nil {
						return err
					}
					// Start a new pure storage row
					row = &Dump{
						Height: endHeight,
						AccountStorage: &AccountStorage{
							Address: acc.Address,
							Storage: make([]*Storage, 0),
						},
					}
				}
				row.AccountStorage.Storage = append(row.AccountStorage.Storage, &Storage{Key: key, Value: value})
				storageBytes += len(key) + len(value)
				return nil
			})
			if err != nil {
				return err
			}

			// Don't send empty storage
			if len(row.AccountStorage.Storage) == 0 {
				row.AccountStorage = nil
				// Don't send an empty row
				if row.Account == nil {
					// We started a new storage row, but there was no subsequent storage to go in it
					return nil
				}
			}

			err = sink.Send(row)
			if err != nil {
				return err
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
			return sink.Send(&Dump{Height: endHeight, Name: entry})
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
					err := sink.Send(&Dump{Height: ev.Event.Header.Height, EVMEvent: &evmevent})
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

// Return a Source that is a Pipe fed from this Dumper's Transmit function
func (ds *Dumper) Source(startHeight, endHeight uint64, options Option) Source {
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

func (ds *Dumper) WithLogger(logger *logging.Logger) *Dumper {
	ds.logger = logger
	return ds
}

// Write a dump to the Writer out by pulling rows from stream
func Write(out io.Writer, source Source, useBinaryEncoding bool, options Option) error {
	st := state.NewState(db.NewMemDB())
	_, _, err := st.Update(func(ws state.Updatable) error {
		for {
			resp, err := source.Recv()
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

			if useBinaryEncoding {
				_, err := encoding.WriteMessage(out, resp)
				if err != nil {
					return fmt.Errorf("failed write to binary dump message: %v", err)
				}
				return nil
			}

			bs, err := json.Marshal(resp)
			if err != nil {
				return fmt.Errorf("failed to marshall dump: %v", err)
			}

			if len(bs) > 0 {
				bs = append(bs, []byte("\n")...)
				n, err := out.Write(bs)
				if err == nil && n < len(bs) {
					return fmt.Errorf("failed to write dump: %v", err)
				}
			}
		}

		return nil
	})
	return err
}
