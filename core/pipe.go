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

package core

import (
  types "github.com/eris-ltd/eris-db/core/types"
)

// TODO: [ben] This respects the old Pipe interface from Eris-DB.
// This made sense as a wrapper around the old Tendermint, but now
// it strongly reflects the internal details of old Tendermint outwards
// and provides little value as an abstraction.
// The refactor needed here for eris-db-0.12.1 is to expose a language
// of transactions, block verification and accounts, grouping
// these interfaces into an Engine, Communicator, NameReg, Permissions (suggestion)

import (
  events           "github.com/tendermint/go-events"
  tendermint_types "github.com/tendermint/tendermint/types"

  transactions "github.com/eris-ltd/eris-db/txs"
)

type Pipe interface {
  Accounts() Accounts
  Blockchain() Blockchain
  Consensus() Consensus
  Events() EventEmitter
  NameReg() NameReg
  Net() Net
  Transactor() Transactor
}

type Accounts interface {
  GenPrivAccount() (*types.PrivAccount, error)
  GenPrivAccountFromKey(privKey []byte) (*types.PrivAccount, error)
  Accounts([]*types.FilterData) (*types.AccountList, error)
  Account(address []byte) (*types.Account, error)
  Storage(address []byte) (*types.Storage, error)
  StorageAt(address, key []byte) (*types.StorageItem, error)
}

type Blockchain interface {
  Info() (*types.BlockchainInfo, error)
  GenesisHash() ([]byte, error)
  ChainId() (string, error)
  LatestBlockHeight() (int, error)
  LatestBlock() (*tendermint_types.Block, error)
  Blocks([]*types.FilterData) (*types.Blocks, error)
  Block(height int) (*tendermint_types.Block, error)
}

type Consensus interface {
  State() (*types.ConsensusState, error)
  Validators() (*types.ValidatorList, error)
}

type EventEmitter interface {
  Subscribe(subId, event string, callback func(events.EventData)) (bool, error)
  Unsubscribe(subId string) (bool, error)
}

type NameReg interface {
  Entry(key string) (*transactions.NameRegEntry, error)
  Entries([]*types.FilterData) (*types.ResultListNames, error)
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
  BroadcastTx(tx transactions.Tx) (*types.Receipt, error)
  Transact(privKey, address, data []byte, gasLimit,
    fee int64) (*types.Receipt, error)
  TransactAndHold(privKey, address, data []byte, gasLimit,
    fee int64) (*transactions.EventDataCall, error)
  TransactNameReg(privKey []byte, name, data string, amount,
    fee int64) (*types.Receipt, error)
  UnconfirmedTxs() (*types.UnconfirmedTxs, error)
  SignTx(tx transactions.Tx,
    privAccounts []*types.PrivAccount) (transactions.Tx, error)
}
