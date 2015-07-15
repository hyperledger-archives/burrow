// The pipe is used to call methods on the Tendermint node.
package pipe

import (
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/account"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/node"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/types"
)

type (

	// Main interface for the pipe. Things here are pretty self-evident.
	Pipe interface {
		Accounts() Accounts
		Blockchain() Blockchain
		Consensus() Consensus
		Events() EventEmitter
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
		Subscribe(subId, event string, callback func(interface{})) (bool, error)
		Unsubscribe(subId string) (bool, error)
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
		Call(address, data []byte) (*Call, error)
		CallCode(code, data []byte) (*Call, error)
		BroadcastTx(tx types.Tx) (*Receipt, error)
		Transact(privKey, address, data []byte, gasLimit, fee int64) (*Receipt, error)
		UnconfirmedTxs() (*UnconfirmedTxs, error)
		SignTx(tx types.Tx, privAccounts []*account.PrivAccount) (types.Tx, error)
	}
)

// Base struct for getting rpc proxy objects (node.Node has no interface).
type PipeImpl struct {
	tNode      *node.Node
	accounts   Accounts
	blockchain Blockchain
	consensus  Consensus
	events     EventEmitter
	net        Net
	txs        Transactor
}

// Create a new rpc pipe.
func NewPipe(tNode *node.Node) Pipe {
	accounts := newAccounts(tNode.ConsensusState(), tNode.MempoolReactor())
	blockchain := newBlockchain(tNode.BlockStore())
	consensus := newConsensus(tNode.ConsensusState(), tNode.Switch())
	events := newEvents(tNode.EventSwitch())
	net := newNetwork(tNode.Switch())
	txs := newTransactor(tNode.EventSwitch(), tNode.ConsensusState(), tNode.MempoolReactor(), events)
	return &PipeImpl{
		tNode,
		accounts,
		blockchain,
		consensus,
		events,
		net,
		txs,
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

func (this *PipeImpl) Net() Net {
	return this.net
}

func (this *PipeImpl) Transactor() Transactor {
	return this.txs
}
