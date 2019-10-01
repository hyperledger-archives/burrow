package tendermint

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/consensus/abci"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/proxy"
	tmTypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
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
	app *abci.App, metricsProvider node.MetricsProvider, marmotNodeKey crypto.PrivateKey, logger *logging.Logger) (*Node, error) {

	var err error
	// disable Tendermint's RPC
	conf.RPC.ListenAddress = ""

	if marmotNodeKey.CurveType != crypto.CurveTypeEd25519 {
		return nil, fmt.Errorf("tendermint node key must be ed25519")
	}

	var pkey ed25519.PrivKeyEd25519
	copy(pkey[:], marmotNodeKey.PrivateKey)
	nodeKey := &p2p.NodeKey{PrivKey: pkey}

	nde := &Node{}
	nde.Node, err = node.NewNode(conf, privValidator,
		nodeKey, proxy.NewLocalClientCreator(app),
		func() (*tmTypes.GenesisDoc, error) {
			return genesisDoc, nil
		},
		nde.DBProvider,
		metricsProvider,
		NewLogger(logger.WithPrefix(structure.ComponentKey, structure.Tendermint).
			With(structure.ScopeKey, "tendermint.NewNode")))
	if err != nil {
		return nil, err
	}
	app.SetMempoolLocker(nde.Mempool())
	return nde, nil
}

func DeriveGenesisDoc(burrowGenesisDoc *genesis.GenesisDoc, appHash []byte) *tmTypes.GenesisDoc {
	validators := make([]tmTypes.GenesisValidator, len(burrowGenesisDoc.Validators))
	for i, validator := range burrowGenesisDoc.Validators {
		validators[i] = tmTypes.GenesisValidator{
			Address: validator.PublicKey.TendermintAddress(),
			PubKey:  validator.PublicKey.TendermintPubKey(),
			Name:    validator.Name,
			Power:   int64(validator.Amount),
		}
	}
	consensusParams := tmTypes.DefaultConsensusParams()
	// This is the smallest increment we can use to get a strictly increasing sequence
	// of block time - we set it low to avoid skew
	// if the BlockTimeIota is longer than the average block time
	consensusParams.Block.TimeIotaMs = 1

	return &tmTypes.GenesisDoc{
		ChainID:         burrowGenesisDoc.ChainID(),
		GenesisTime:     burrowGenesisDoc.GenesisTime,
		Validators:      validators,
		AppHash:         appHash,
		ConsensusParams: consensusParams,
	}
}

func NewNodeInfo(ni p2p.DefaultNodeInfo) *NodeInfo {
	address, _ := crypto.AddressFromHexString(string(ni.ID()))
	return &NodeInfo{
		ID:            address,
		Moniker:       ni.Moniker,
		ListenAddress: ni.ListenAddr,
		Version:       ni.Version,
		Channels:      binary.HexBytes(ni.Channels),
		Network:       ni.Network,
		RPCAddress:    ni.Other.RPCAddress,
		TxIndex:       ni.Other.TxIndex,
	}
}

func NewNodeKey() *p2p.NodeKey {
	privKey := ed25519.GenPrivKey()
	return &p2p.NodeKey{
		PrivKey: privKey,
	}
}

func WriteNodeKey(nodeKeyFile string, key json.RawMessage) error {
	err := os.MkdirAll(path.Dir(nodeKeyFile), 0777)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(nodeKeyFile, key, 0600)
}

func EnsureNodeKey(nodeKeyFile string) (*p2p.NodeKey, error) {
	err := os.MkdirAll(path.Dir(nodeKeyFile), 0777)
	if err != nil {
		return nil, err
	}

	return p2p.LoadOrGenNodeKey(nodeKeyFile)
}
