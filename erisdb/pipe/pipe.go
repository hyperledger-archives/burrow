// This
package pipe

import (
	"github.com/tendermint/tendermint/account"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/types"
)

type (

	// Main interface for the pipe. Things here are pretty self-evident.
	Pipe interface {
		Accounts() Accounts
		Blockchain() Blockchain
		Consensus() Consensus
		Events() Events
		Net() Net
		Txs() Txs
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
		LatestBlockHeight() (uint, error)
		LatestBlock() (*types.Block, error)
		Blocks([]*FilterData) (*Blocks, error)
		Block(height uint) (*types.Block, error)
	}

	Consensus interface {
		State() (*ConsensusState, error)
		Validators() (*ValidatorList, error)
	}

	Events interface {
		Subscribe(subId, event string, callback func(interface{})) (bool, error)
		Unsubscribe(subId string) (bool, error)
	}

	Net interface {
		Info() (*NetworkInfo, error)
		Moniker() (string, error)
		Listening() (bool, error)
		Listeners() ([]string, error)
		Peers() ([]*Peer, error)
		Peer(string) (*Peer, error)
	}

	Txs interface {
		Call(address, data []byte) (*Call, error)
		CallCode(code, data []byte) (*Call, error)
		BroadcastTx(tx types.Tx) (*Receipt, error)
		Transact(privKey, address, data []byte, gasLimit, fee uint64) (*Receipt, error)
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
	events     Events
	net        Net
	txs        Txs
}

// Create a new rpc pipe.
func NewPipe(tNode *node.Node) Pipe {
	accounts := newAccounts(tNode.ConsensusState(), tNode.MempoolReactor())
	blockchain := newBlockchain(tNode.BlockStore())
	consensus := newConsensus(tNode.ConsensusState(), tNode.Switch())
	events := newEvents(tNode.EventSwitch())
	net := newNetwork(tNode.Switch())
	txs := newTxs(tNode.ConsensusState(), tNode.MempoolReactor())
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

func (this *PipeImpl) Events() Events {
	return this.events
}

func (this *PipeImpl) Net() Net {
	return this.net
}

func (this *PipeImpl) Txs() Txs {
	return this.txs
}
