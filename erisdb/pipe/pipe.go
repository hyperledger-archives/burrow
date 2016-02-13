// The pipe is used to call methods on the Tendermint node.
package pipe

import (
	"github.com/eris-ltd/eris-db/account"
	em "github.com/tendermint/go-events"
	"github.com/tendermint/tendermint/types"

	"github.com/eris-ltd/eris-db/tmsp"
	txs "github.com/eris-ltd/eris-db/txs"
)

type (

	// Main interface for the pipe. Things here are pretty self-evident.
	Pipe interface {
		Accounts() Accounts
		Blockchain() Blockchain
		Consensus() Consensus
		Events() EventEmitter
		NameReg() NameReg
		Net() Net
		Transactor() Transactor
	}

	Accounts interface {
		GenPrivAccount() (*account.PrivAccount, error)
		GenPrivAccountFromKey(privKey []byte) (*account.PrivAccount, error)
		Accounts([]*FilterData) (*AccountList, error)
		Account(address []byte) (*account.Account, error)
		Storage(address []byte) (*Storage, error)
		StorageAt(address, key []byte) (*StorageItem, error)
	}

	Blockchain interface {
		Info() (*BlockchainInfo, error)
		GenesisHash() ([]byte, error)
		ChainId() (string, error)
		LatestBlockHeight() (int, error)
		LatestBlock() (*types.Block, error)
		Blocks([]*FilterData) (*Blocks, error)
		Block(height int) (*types.Block, error)
	}

	Consensus interface {
		State() (*ConsensusState, error)
		Validators() (*ValidatorList, error)
	}

	EventEmitter interface {
		Subscribe(subId, event string, callback func(em.EventData)) (bool, error)
		Unsubscribe(subId string) (bool, error)
	}

	NameReg interface {
		Entry(key string) (*txs.NameRegEntry, error)
		Entries([]*FilterData) (*ResultListNames, error)
	}

	Net interface {
		Info() (*NetworkInfo, error)
		ClientVersion() (string, error)
		Moniker() (string, error)
		Listening() (bool, error)
		Listeners() ([]string, error)
		Peers() ([]*Peer, error)
		Peer(string) (*Peer, error)
	}

	Transactor interface {
		Call(fromAddress, toAddress, data []byte) (*Call, error)
		CallCode(fromAddress, code, data []byte) (*Call, error)
		BroadcastTx(tx txs.Tx) (*Receipt, error)
		Transact(privKey, address, data []byte, gasLimit, fee int64) (*Receipt, error)
		TransactAndHold(privKey, address, data []byte, gasLimit, fee int64) (*txs.EventDataCall, error)
		TransactNameReg(privKey []byte, name, data string, amount, fee int64) (*Receipt, error)
		UnconfirmedTxs() (*UnconfirmedTxs, error)
		SignTx(tx txs.Tx, privAccounts []*account.PrivAccount) (txs.Tx, error)
	}
)

// Base struct for getting rpc proxy objects (node.Node has no interface).
type PipeImpl struct {
	//tNode      *node.Node
	erisdbApp  *tmsp.ErisDBApp
	accounts   Accounts
	blockchain Blockchain
	consensus  Consensus
	events     EventEmitter
	namereg    NameReg
	net        Net
	txs        Transactor
}

// Create a new rpc pipe.
func NewPipe(erisdbApp *tmsp.ErisDBApp) Pipe {
	eventSwitch := em.NewEventSwitch()
	events := newEvents(eventSwitch)

	accounts := newAccounts(erisdbApp)
	namereg := newNamereg(erisdbApp)
	txs := newTransactor(eventSwitch, erisdbApp, events)

	// TODO: make interface to tendermint core's rpc for these
	// blockchain := newBlockchain(blockStore)
	// consensus := newConsensus(erisdbApp)
	// net := newNetwork(erisdbApp)

	return &PipeImpl{
		erisdbApp: erisdbApp,
		accounts:  accounts,
		// blockchain,
		// consensus,
		events:  events,
		namereg: namereg,
		// net,
		txs: txs,
	}
}

func (this *PipeImpl) Accounts() Accounts {
	return this.accounts
}

func (this *PipeImpl) Blockchain() Blockchain {
	return this.blockchain
}

func (this *PipeImpl) Consensus() Consensus {
	return this.consensus
}

func (this *PipeImpl) Events() EventEmitter {
	return this.events
}

func (this *PipeImpl) NameReg() NameReg {
	return this.namereg
}

func (this *PipeImpl) Net() Net {
	return this.net
}

func (this *PipeImpl) Transactor() Transactor {
	return this.txs
}
