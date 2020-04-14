package dump

import (
	bin "encoding/binary"
	"fmt"
	"io"
	"math/rand"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/execution/native"
	"github.com/hyperledger/burrow/genesis"
)

type MockSource struct {
	Accounts   int
	MaxStorage int
	Names      int
	Events     int
	*Mockchain
	rand *rand.Rand
}

var _ Source = &MockSource{}

func NewMockSource(accounts, maxStorage, names, events int) *MockSource {
	return &MockSource{
		Accounts:   accounts,
		MaxStorage: maxStorage,
		Names:      names,
		Events:     events,
		Mockchain:  NewMockchain("Mockchain", 0),
		rand:       rand.New(rand.NewSource(2323524)),
	}
}

func (m *MockSource) Recv() (*Dump, error) {
	row := Dump{Height: m.LastBlockHeight()}

	// In order to create the same state as from a real dump we need to honour the dump order:
	// [accounts[storage...]...][names...][events...]
	if m.Accounts > 0 {
		var addr crypto.Address
		bin.BigEndian.PutUint64(addr[:], uint64(m.Accounts))

		row.Account = &acm.Account{
			Address: addr,
			Balance: m.rand.Uint64(),
		}

		if m.Accounts%2 > 0 {
			row.Account.EVMCode = make([]byte, m.rand.Intn(10000))
			m.rand.Read(row.Account.EVMCode)
			row.Account.EVMOpcodeBitset = native.EVMOpcodeBitset(row.Account.EVMCode)
		} else {
			row.Account.PublicKey = crypto.PublicKey{}
		}
		m.Accounts--
		if m.MaxStorage > 0 {
			// We don't send empty storage
			storagelen := 1 + m.rand.Intn(m.MaxStorage)

			row.AccountStorage = &AccountStorage{
				Address: addr,
				Storage: make([]*Storage, storagelen),
			}

			for i := 0; i < storagelen; i++ {
				var key binary.Word256
				// Put account index in first 8 bytes
				copy(key[:8], addr[:8])
				// Put storage index in last 8 bytes
				bin.BigEndian.PutUint64(key[24:], uint64(i))
				row.AccountStorage.Storage[i] = &Storage{Key: key, Value: key[:]}
			}
		}
	} else if m.Accounts == 0 {
		// Finally send the global permissions account (makes for easier equality checks with genesis state dump)
		row.Account = genesis.DefaultPermissionsAccount
		m.Accounts--
	} else if m.Names > 0 {
		row.Name = &names.Entry{
			Name:    fmt.Sprintf("name%d", m.Names),
			Data:    fmt.Sprintf("data%x", m.Names),
			Owner:   crypto.ZeroAddress,
			Expires: 1337,
		}
		m.Names--
	} else if m.Events > 0 {
		datalen := 1 + m.rand.Intn(10)
		data := make([]byte, datalen*32)
		topiclen := 1 + m.rand.Intn(5)
		topics := make([]binary.Word256, topiclen)
		row.EVMEvent = &EVMEvent{
			ChainID: m.ChainID(),
			Event: &exec.LogEvent{
				Address: crypto.ZeroAddress,
				Data:    data,
				Topics:  topics,
			},
		}
		m.Events--
	} else {
		return nil, io.EOF
	}

	return &row, nil
}

type Mockchain struct {
	chainID         string
	lastBlockHeight uint64
}

var _ Blockchain = &Mockchain{}

func NewMockchain(chainID string, lastBlockHeight uint64) *Mockchain {
	return &Mockchain{
		chainID:         chainID,
		lastBlockHeight: lastBlockHeight,
	}
}

func (mc *Mockchain) ChainID() string {
	return mc.chainID
}

func (mc *Mockchain) LastBlockHeight() uint64 {
	return mc.lastBlockHeight
}
