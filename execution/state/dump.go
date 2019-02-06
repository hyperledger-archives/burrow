package state

import (
	"encoding/json"
	"io"
	"os"

	amino "github.com/tendermint/go-amino"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/dump"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/txs/payload"
)

func (s *State) LoadDump(filename string) error {
	cdc := amino.NewCodec()
	f, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}

	txs := make([]*exec.TxExecution, 0)

	var tx *exec.TxExecution

	apply := func(row dump.Dump) error {
		if row.Account != nil {
			if row.Account.Address != acm.GlobalPermissionsAddress {
				return s.writeState.UpdateAccount(row.Account)
			}
		}
		if row.AccountStorage != nil {
			for _, storage := range row.AccountStorage.Storage {
				err := s.writeState.SetStorage(row.AccountStorage.Address, storage.Key, storage.Value)
				if err != nil {
					return err
				}
			}
		}
		if row.Name != nil {
			return s.writeState.UpdateName(row.Name)
		}
		if row.EVMEvent != nil {

			if tx != nil && row.Height != tx.Height {
				txs = append(txs, tx)
				tx = nil
			}
			if tx == nil {
				tx = &exec.TxExecution{
					TxType: payload.TypeCall,
					TxHash: make([]byte, 32),
					Height: row.Height,
					Origin: &exec.Origin{
						ChainID: row.EVMEvent.ChainID,
						Time:    row.EVMEvent.Time,
					},
				}
			}

			tx.Events = append(tx.Events, &exec.Event{
				Header: &exec.Header{
					TxType:    payload.TypeCall,
					EventType: exec.TypeLog,
					Height:    row.Height,
				},
				Log: row.EVMEvent.Event,
			})
		}
		return nil
	}

	if tx != nil {
		txs = append(txs, tx)
	}

	// first try amino
	first := true

	for err == nil {
		var row dump.Dump

		_, err = cdc.UnmarshalBinaryLengthPrefixedReader(f, &row, 0)
		if err != nil {
			break
		}

		first = false
		err = apply(row)
	}

	// if we failed at the first row, try json
	if err != io.EOF && first {
		err = nil
		f.Seek(0, 0)

		decoder := json.NewDecoder(f)

		for err == nil {
			var row dump.Dump

			err = decoder.Decode(&row)
			if err != nil {
				break
			}

			err = apply(row)
		}
	}

	errAddTxs := s.writeState.AddBlock(&exec.BlockExecution{
		Height:       0,
		TxExecutions: txs,
	})
	if errAddTxs != nil {
		return errAddTxs
	}

	if err == io.EOF {
		return nil
	}

	return err
}

func (s *State) Dump() string {
	return s.writeState.forest.Dump()
}
