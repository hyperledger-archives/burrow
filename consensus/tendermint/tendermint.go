package tendermint

import (
	"os"
	"path"

	bcm "github.com/hyperledger/burrow/blockchain"
	"github.com/hyperledger/burrow/consensus/tendermint/abci"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/txs"
	"github.com/tendermint/tendermint/config"
	tmCrypto "github.com/tendermint/tendermint/crypto"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/proxy"
	tmTypes "github.com/tendermint/tendermint/types"
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

func NewNode(conf *config.Config, privValidator tmTypes.PrivValidator, genesisDoc *tmTypes.GenesisDoc,
	blockchain *bcm.Blockchain, checker execution.BatchExecutor, committer execution.BatchCommitter,
	txDecoder txs.Decoder, metricsProvider node.MetricsProvider, panicFunc func(error), logger *logging.Logger) (*Node, error) {

	var err error
	// disable Tendermint's RPC
	conf.RPC.ListenAddress = ""

	err = os.MkdirAll(path.Dir(conf.NodeKeyFile()), 0777)
	if err != nil {
		return nil, err
	}

	nde := &Node{}
	app := abci.NewApp(blockchain, checker, committer, txDecoder, panicFunc, logger)
	conf.NodeKeyFile()
	nde.Node, err = node.NewNode(conf, privValidator,
		proxy.NewLocalClientCreator(app),
		func() (*tmTypes.GenesisDoc, error) {
			return genesisDoc, nil
		},
		nde.DBProvider,
		metricsProvider,
		NewLogger(logger.WithPrefix(structure.ComponentKey, "Tendermint").
			With(structure.ScopeKey, "tendermint.NewNode")))
	if err != nil {
		return nil, err
	}
	app.SetMempoolLocker(nde.MempoolReactor().Mempool)
	return nde, nil
}

func DeriveGenesisDoc(burrowGenesisDoc *genesis.GenesisDoc) *tmTypes.GenesisDoc {
	validators := make([]tmTypes.GenesisValidator, len(burrowGenesisDoc.Validators))
	for i, validator := range burrowGenesisDoc.Validators {
		tm := tmCrypto.PubKeyEd25519{}
		copy(tm[:], validator.PublicKey.RawBytes())
		validators[i] = tmTypes.GenesisValidator{
			PubKey: tm,
			Name:   validator.Name,
			Power:  int64(validator.Amount),
		}
	}
	return &tmTypes.GenesisDoc{
		ChainID:         burrowGenesisDoc.ChainID(),
		GenesisTime:     burrowGenesisDoc.GenesisTime,
		Validators:      validators,
		AppHash:         burrowGenesisDoc.Hash(),
		ConsensusParams: tmTypes.DefaultConsensusParams(),
	}
}

func NewBlockEvent(message interface{}) *tmTypes.EventDataNewBlock {
	tmEventData, ok := message.(tmTypes.TMEventData)
	if ok {
		eventDataNewBlock, ok := tmEventData.(tmTypes.EventDataNewBlock)
		if ok {
			return &eventDataNewBlock
		}
	}
	return nil
}
