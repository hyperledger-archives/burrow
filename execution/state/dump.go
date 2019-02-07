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

var cdc = amino.NewCodec()

type DumpReader interface {
	Next() (*dump.Dump, error)
}

type FileDumpReader struct {
	file    *os.File
	decoder *json.Decoder
}

func NewFileDumpReader(filename string) (DumpReader, error) {
	f, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &FileDumpReader{file: f}, nil
}

func (f *FileDumpReader) Next() (*dump.Dump, error) {
	var row dump.Dump
	var err error

	if f.decoder != nil {
		err = f.decoder.Decode(&row)
	} else {
		_, err = cdc.UnmarshalBinaryLengthPrefixedReader(f.file, &row, 0)

		if err != nil && err != io.EOF && f.decoder == nil {
			f.file.Seek(0, 0)

			f.decoder = json.NewDecoder(f.file)

			return f.Next()
		}
	}

	if err == io.EOF {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &row, err
}

func (s *State) LoadDump(reader DumpReader) error {
	txs := make([]*exec.TxExecution, 0)

	var tx *exec.TxExecution

	for {
		row, err := reader.Next()

		if err != nil {
			return err
		}

		if row == nil {
			break
		}

		if row.Account != nil {
			if row.Account.Address != acm.GlobalPermissionsAddress {
				s.writeState.UpdateAccount(row.Account)
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
			s.writeState.UpdateName(row.Name)
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
	}

	if tx != nil {
		txs = append(txs, tx)
	}

	return s.writeState.AddBlock(&exec.BlockExecution{
		Height:       0,
		TxExecutions: txs,
	})
}

func (s *State) Dump() string {
	return s.writeState.forest.Dump()
}
