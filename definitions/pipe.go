// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package definitions

// TODO: [ben] This respects the old Pipe interface from Eris-DB.
// This made sense as a wrapper around the old Tendermint, but now
// it strongly reflects the internal details of old Tendermint outwards
// and provides little value as an abstraction.
// The refactor needed here for eris-db-0.12.1 is to expose a language
// of transactions, block verification and accounts, grouping
// these interfaces into an Engine, Communicator, NameReg, Permissions (suggestion)

import (
	account "github.com/monax/burrow/account"
	blockchain_types "github.com/monax/burrow/blockchain/types"
	consensus_types "github.com/monax/burrow/consensus/types"
	core_types "github.com/monax/burrow/core/types"
	types "github.com/monax/burrow/core/types"
	event "github.com/monax/burrow/event"
	logging_types "github.com/monax/burrow/logging/types"
	manager_types "github.com/monax/burrow/manager/types"
	"github.com/monax/burrow/txs"
)

type Pipe interface {
	Accounts() Accounts
	Blockchain() blockchain_types.Blockchain
	Events() event.EventEmitter
	NameReg() NameReg
	Transactor() Transactor
	// Hash of Genesis state
	GenesisHash() []byte
	Logger() logging_types.InfoTraceLogger
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
	Send(privKey, toAddress []byte, amount int64) (*txs.Receipt, error)
	SendAndHold(privKey, toAddress []byte, amount int64) (*txs.Receipt, error)
	TransactNameReg(privKey []byte, name, data string, amount,
		fee int64) (*txs.Receipt, error)
	SignTx(tx txs.Tx, privAccounts []*account.PrivAccount) (txs.Tx, error)
}
