package tendermint

import (
	"fmt"

	"context"

	"os"
	"path"

	bcm "github.com/hyperledger/burrow/blockchain"
	"github.com/hyperledger/burrow/consensus/tendermint/abci"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/txs"
	abci_types "github.com/tendermint/abci/types"
	tm_crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/proxy"
	tm_types "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tmlibs/db"
)

// Serves as a wrapper around the Tendermint node's closeable resources (database connections)
type Node struct {
	*node.Node
	closers []interface {
		Close()
	}
}

func DBProvider(ID string, backendType dbm.DBBackendType, dbDir string) dbm.DB {
	return dbm.NewDB(ID, backendType, dbDir)
}

// Since Tendermint doesn't close its DB connections
func (n *Node) DBProvider(ctx *node.DBContext) (dbm.DB, error) {
	db := DBProvider(ctx.ID, dbm.DBBackendType(ctx.Config.DBBackend), ctx.Config.DBDir())
	n.closers = append(n.closers, db)
	return db, nil
}

func (n *Node) Close() {
	for _, closer := range n.closers {
		closer.Close()
	}
}

func NewNode(
	conf *config.Config,
	privValidator tm_types.PrivValidator,
	genesisDoc *tm_types.GenesisDoc,
	blockchain bcm.MutableBlockchain,
	checker execution.BatchExecutor,
	committer execution.BatchCommitter,
	logger *logging.Logger) (*Node, error) {

	var err error
	// disable Tendermint's RPC
	conf.RPC.ListenAddress = ""

	err = os.MkdirAll(path.Dir(conf.NodeKeyFile()), 0777)
	if err != nil {
		return nil, err
	}

	nde := &Node{}
	app := abci.NewApp(blockchain, checker, committer, logger)
	conf.NodeKeyFile()
	nde.Node, err = node.NewNode(conf, privValidator,
		proxy.NewLocalClientCreator(app),
		func() (*tm_types.GenesisDoc, error) {
			return genesisDoc, nil
		},
		nde.DBProvider,
		NewLogger(logger.WithPrefix(structure.ComponentKey, "Tendermint").
			With(structure.ScopeKey, "tendermint.NewNode")))
	if err != nil {
		return nil, err
	}
	app.SetMempoolLocker(nde.MempoolReactor().Mempool)
	return nde, nil
}

func BroadcastTxAsyncFunc(validator *Node, txEncoder txs.Encoder) func(tx txs.Tx,
	callback func(res *abci_types.Response)) error {

	return func(tx txs.Tx, callback func(res *abci_types.Response)) error {
		txBytes, err := txEncoder.EncodeTx(tx)
		if err != nil {
			return fmt.Errorf("error encoding transaction: %v", err)
		}

		err = validator.MempoolReactor().BroadcastTx(txBytes, callback)
		if err != nil {
			return fmt.Errorf("error broadcasting transaction: %v", err)
		}
		return nil
	}
}

func DeriveGenesisDoc(burrowGenesisDoc *genesis.GenesisDoc) *tm_types.GenesisDoc {
	validators := make([]tm_types.GenesisValidator, len(burrowGenesisDoc.Validators))
	for i, validator := range burrowGenesisDoc.Validators {
		tm := tm_crypto.PubKeyEd25519{}
		copy(tm[:], validator.PublicKey.RawBytes())
		validators[i] = tm_types.GenesisValidator{
			PubKey: tm_crypto.PubKey{tm},
			Name:   validator.Name,
			Power:  int64(validator.Amount),
		}
	}
	return &tm_types.GenesisDoc{
		ChainID:     burrowGenesisDoc.ChainID(),
		GenesisTime: burrowGenesisDoc.GenesisTime,
		Validators:  validators,
		AppHash:     burrowGenesisDoc.Hash(),
	}
}

func SubscribeNewBlock(ctx context.Context, subscribable event.Subscribable, subscriber string,
	ch chan<- *tm_types.EventDataNewBlock) error {
	query := event.QueryForEventID(tm_types.EventNewBlock)

	return event.SubscribeCallback(ctx, subscribable, subscriber, query, func(message interface{}) bool {
		tmEventData, ok := message.(tm_types.TMEventData)
		if ok {
			eventDataNewBlock, ok := tmEventData.Unwrap().(tm_types.EventDataNewBlock)
			if ok {
				ch <- &eventDataNewBlock
			}
		}
		return true
	})
}
