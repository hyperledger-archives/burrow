// Copyright 2015, 2016 Eris Industries (UK) Ltd.
// This file is part of Eris-RT

// Eris-RT is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Eris-RT is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Eris-RT.  If not, see <http://www.gnu.org/licenses/>.

package definitions

// TODO: [ben] This respects the old Pipe interface from Eris-DB.
// This made sense as a wrapper around the old Tendermint, but now
// it strongly reflects the internal details of old Tendermint outwards
// and provides little value as an abstraction.
// The refactor needed here for eris-db-0.12.1 is to expose a language
// of transactions, block verification and accounts, grouping
// these interfaces into an Engine, Communicator, NameReg, Permissions (suggestion)

import (
	account "github.com/eris-ltd/eris-db/account"
	blockchain_types "github.com/eris-ltd/eris-db/blockchain/types"
	consensus_types "github.com/eris-ltd/eris-db/consensus/types"
	core_types "github.com/eris-ltd/eris-db/core/types"
	types "github.com/eris-ltd/eris-db/core/types"
	event "github.com/eris-ltd/eris-db/event"
	manager_types "github.com/eris-ltd/eris-db/manager/types"
	"github.com/eris-ltd/eris-db/txs"
)

type Pipe interface {
	Accounts() Accounts
	Blockchain() blockchain_types.Blockchain
	Consensus() consensus_types.ConsensusEngine
	Events() event.EventEmitter
	NameReg() NameReg
	Net() Net
	Transactor() Transactor
	// Hash of Genesis state
	GenesisHash() []byte
	// NOTE: [ben] added to Pipe interface on 0.12 refactor
	GetApplication() manager_types.Application
	SetConsensusEngine(consensusEngine consensus_types.ConsensusEngine) error
	GetConsensusEngine() consensus_types.ConsensusEngine
	SetBlockchain(blockchain blockchain_types.Blockchain) error
	GetBlockchain() blockchain_types.Blockchain
	// Support for Tendermint RPC
	GetTendermintPipe() (TendermintPipe, error)
}

type Accounts interface {
	GenPrivAccount() (*account.PrivAccount, error)
	GenPrivAccountFromKey(privKey []byte) (*account.PrivAccount, error)
	Accounts([]*event.FilterData) (*types.AccountList, error)
	Account(address []byte) (*account.Account, error)
	Storage(address []byte) (*types.Storage, error)
	StorageAt(address, key []byte) (*types.StorageItem, error)
}

type NameReg interface {
	Entry(key string) (*core_types.NameRegEntry, error)
	Entries([]*event.FilterData) (*types.ResultListNames, error)
}

type Net interface {
	Info() (*types.NetworkInfo, error)
	ClientVersion() (string, error)
	Moniker() (string, error)
	Listening() (bool, error)
	Listeners() ([]string, error)
	Peers() ([]*types.Peer, error)
	Peer(string) (*types.Peer, error)
}

type Transactor interface {
	Call(fromAddress, toAddress, data []byte) (*types.Call, error)
	CallCode(fromAddress, code, data []byte) (*types.Call, error)
	// Send(privKey, toAddress []byte, amount int64) (*types.Receipt, error)
	// SendAndHold(privKey, toAddress []byte, amount int64) (*types.Receipt, error)
	BroadcastTx(tx txs.Tx) (*txs.Receipt, error)
	Transact(privKey, address, data []byte, gasLimit,
		fee int64) (*txs.Receipt, error)
	TransactAndHold(privKey, address, data []byte, gasLimit,
		fee int64) (*txs.EventDataCall, error)
	TransactNameReg(privKey []byte, name, data string, amount,
		fee int64) (*txs.Receipt, error)
	UnconfirmedTxs() (*txs.UnconfirmedTxs, error)
	SignTx(tx txs.Tx, privAccounts []*account.PrivAccount) (txs.Tx, error)
}
