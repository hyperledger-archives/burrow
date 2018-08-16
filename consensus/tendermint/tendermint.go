package tendermint

import (
	"os"
	"path"

	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/consensus/tendermint/abci"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/tendermint/tendermint/config"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
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
	app *abci.App, metricsProvider node.MetricsProvider, logger *logging.Logger) (*Node, error) {

	var err error
	// disable Tendermint's RPC
	conf.RPC.ListenAddress = ""

	err = os.MkdirAll(path.Dir(conf.NodeKeyFile()), 0777)
	if err != nil {
		return nil, err
	}

	nde := &Node{}
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
		validators[i] = tmTypes.GenesisValidator{
			PubKey: validator.PublicKey.TendermintPubKey(),
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

func NewNodeInfo(ni p2p.NodeInfo) *NodeInfo {
	address, _ := crypto.AddressFromHexString(string(ni.ID))
	return &NodeInfo{
		ID:            address,
		Moniker:       ni.Moniker,
		ListenAddress: ni.ListenAddr,
		Version:       ni.Version,
		Channels:      binary.HexBytes(ni.Channels),
		Network:       ni.Network,
		Other:         ni.Other,
	}
}
