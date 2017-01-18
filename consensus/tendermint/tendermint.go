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

// version provides the current Eris-DB version and a VersionIdentifier
// for the modules to identify their version with.

package tendermint

import (
	"fmt"
	"path"
	"strings"
	"sync"

	crypto "github.com/tendermint/go-crypto"
	p2p "github.com/tendermint/go-p2p"
	tendermint_consensus "github.com/tendermint/tendermint/consensus"
	node "github.com/tendermint/tendermint/node"
	proxy "github.com/tendermint/tendermint/proxy"
	tendermint_types "github.com/tendermint/tendermint/types"
	tmsp_types "github.com/tendermint/tmsp/types"

	edb_event "github.com/eris-ltd/eris-db/event"

	config "github.com/eris-ltd/eris-db/config"
	manager_types "github.com/eris-ltd/eris-db/manager/types"
	// files  "github.com/eris-ltd/eris-db/files"
	blockchain_types "github.com/eris-ltd/eris-db/blockchain/types"
	consensus_types "github.com/eris-ltd/eris-db/consensus/types"
	"github.com/eris-ltd/eris-db/logging"
	"github.com/eris-ltd/eris-db/logging/loggers"
	"github.com/eris-ltd/eris-db/txs"
	"github.com/tendermint/go-wire"
)

type Tendermint struct {
	tmintNode   *node.Node
	tmintConfig *TendermintConfig
	chainId     string
	logger      loggers.InfoTraceLogger
}

// Compiler checks to ensure Tendermint successfully implements
// eris-db/definitions Consensus and Blockchain
var _ consensus_types.ConsensusEngine = (*Tendermint)(nil)
var _ blockchain_types.Blockchain = (*Tendermint)(nil)

func NewTendermint(moduleConfig *config.ModuleConfig,
	application manager_types.Application,
	logger loggers.InfoTraceLogger) (*Tendermint, error) {
	// re-assert proper configuration for module
	if moduleConfig.Version != GetTendermintVersion().GetMinorVersionString() {
		return nil, fmt.Errorf("Version string %s did not match %s",
			moduleConfig.Version, GetTendermintVersion().GetMinorVersionString())
	}
	// loading the module has ensured the working and data directory
	// for tendermint have been created, but the config files needs
	// to be written in tendermint's root directory.
	// NOTE: [ben] as elsewhere Sub panics if config file does not have this
	// subtree. To shield in go-routine, or PR to viper.
	if !moduleConfig.Config.IsSet("configuration") {
		return nil, fmt.Errorf("Failed to extract Tendermint configuration subtree.")
	}
	tendermintConfigViper, err := config.ViperSubConfig(moduleConfig.Config, "configuration")
	if tendermintConfigViper == nil {
		return nil,
			fmt.Errorf("Failed to extract Tendermint configuration subtree: %s", err)
	}
	// wrap a copy of the viper config in a tendermint/go-config interface
	tmintConfig := GetTendermintConfig(tendermintConfigViper)
	// complete the tendermint configuration with default flags
	tmintConfig.AssertTendermintDefaults(moduleConfig.ChainId,
		moduleConfig.WorkDir, moduleConfig.DataDir, moduleConfig.RootDir)

	privateValidatorFilePath := path.Join(moduleConfig.RootDir,
		moduleConfig.Config.GetString("private_validator_file"))
	if moduleConfig.Config.GetString("private_validator_file") == "" {
		return nil, fmt.Errorf("No private validator file provided.")
	}
	// override tendermint configurations to force consistency with overruling
	// settings
	tmintConfig.AssertTendermintConsistency(moduleConfig,
		privateValidatorFilePath)
	chainId := tmintConfig.GetString("chain_id")

	logging.TraceMsg(logger, "Loaded Tendermint sub-configuration",
		"chainId", chainId,
		"genesisFile", tmintConfig.GetString("genesis_file"),
		"nodeLocalAddress", tmintConfig.GetString("node_laddr"),
		"moniker", tmintConfig.GetString("moniker"),
		"seeds", tmintConfig.GetString("seeds"),
		"fastSync", tmintConfig.GetBool("fast_sync"),
		"rpcLocalAddress", tmintConfig.GetString("rpc_laddr"),
		"databaseDirectory", tmintConfig.GetString("db_dir"),
		"privateValidatorFile", tmintConfig.GetString("priv_validator_file"),
		"privValFile", moduleConfig.Config.GetString("private_validator_file"))

	// TODO: [ben] do not "or Generate Validator keys", rather fail directly
	// TODO: [ben] implement the signer for Private validator over eris-keys
	// TODO: [ben] copy from rootDir to tendermint workingDir;
	privateValidator := tendermint_types.LoadOrGenPrivValidator(
		path.Join(moduleConfig.RootDir,
			moduleConfig.Config.GetString("private_validator_file")))

	// TODO: [Silas] we want to something better than this like not not have it in
	// the config at all, but for now I think it's much safer to make sure we are
	// not running the tendermint RPC as it could lead to unexpected behaviour,
	// not least if we accidentally try to run it on the same address as our own
	if tmintConfig.GetString("rpc_laddr") != "" {
		logging.InfoMsg(logger, "Force disabling Tendermint's native RPC",
			"provided_rpc_laddr", tmintConfig.GetString("rpc_laddr"))
		tmintConfig.Set("rpc_laddr", "")
	}

	newNode := node.NewNode(tmintConfig, privateValidator, func(_, _ string,
		hash []byte) proxy.AppConn {
		return NewLocalClient(new(sync.Mutex), application)
	})

	listener := p2p.NewDefaultListener("tcp", tmintConfig.GetString("node_laddr"),
		tmintConfig.GetBool("skip_upnp"))

	newNode.AddListener(listener)
	// TODO: [ben] delay starting the node to a different function, to hand
	// control over events to Core
	if err := newNode.Start(); err != nil {
		newNode.Stop()
		return nil, fmt.Errorf("Failed to start Tendermint consensus node: %v", err)
	}
	logging.InfoMsg(logger, "Tendermint consensus node started",
		"nodeAddress", tmintConfig.GetString("node_laddr"),
		"transportProtocol", "tcp",
		"upnp", !tmintConfig.GetBool("skip_upnp"),
		"moniker", tmintConfig.GetString("moniker"))

	// If seedNode is provided by config, dial out.
	if tmintConfig.GetString("seeds") != "" {
		seeds := strings.Split(tmintConfig.GetString("seeds"), ",")
		newNode.DialSeeds(seeds)
		logging.TraceMsg(logger, "Tendermint node called seeds",
			"seeds", seeds)
	}

	return &Tendermint{
		tmintNode:   newNode,
		tmintConfig: tmintConfig,
		chainId:     chainId,
		logger:      logger,
	}, nil
}

//------------------------------------------------------------------------------
// Blockchain implementation

func (tendermint *Tendermint) Height() int {
	return tendermint.tmintNode.BlockStore().Height()
}

func (tendermint *Tendermint) BlockMeta(height int) *tendermint_types.BlockMeta {
	return tendermint.tmintNode.BlockStore().LoadBlockMeta(height)
}

func (tendermint *Tendermint) Block(height int) *tendermint_types.Block {
	return tendermint.tmintNode.BlockStore().LoadBlock(height)
}

func (tendermint *Tendermint) ChainId() string {
	return tendermint.chainId
}

// Consensus implementation
func (tendermint *Tendermint) IsListening() bool {
	return tendermint.tmintNode.Switch().IsListening()
}

func (tendermint *Tendermint) Listeners() []p2p.Listener {
	var copyListeners []p2p.Listener
	// copy slice of Listeners
	copy(copyListeners[:], tendermint.tmintNode.Switch().Listeners())
	return copyListeners
}

func (tendermint *Tendermint) Peers() []*consensus_types.Peer {
	p2pPeers := tendermint.tmintNode.Switch().Peers().List()
	peers := make([]*consensus_types.Peer, 0)
	for _, peer := range p2pPeers {
		peers = append(peers, &consensus_types.Peer{
			NodeInfo:   peer.NodeInfo,
			IsOutbound: peer.IsOutbound(),
		})
	}
	return peers
}

func (tendermint *Tendermint) NodeInfo() *p2p.NodeInfo {
	var copyNodeInfo = new(p2p.NodeInfo)
	// call Switch().NodeInfo is not go-routine safe, so copy
	*copyNodeInfo = *tendermint.tmintNode.Switch().NodeInfo()
	tendermint.tmintNode.ConsensusState().GetRoundState()
	return copyNodeInfo
}

func (tendermint *Tendermint) PublicValidatorKey() crypto.PubKey {
	// TODO: [ben] this is abetment, not yet a go-routine safe solution
	var copyPublicValidatorKey crypto.PubKey
	// crypto.PubKey is an interface so copy underlying struct
	publicKey := tendermint.tmintNode.PrivValidator().PubKey
	switch publicKey.(type) {
	case crypto.PubKeyEd25519:
		// type crypto.PubKeyEd25519 is [32]byte
		copyKeyBytes := publicKey.(crypto.PubKeyEd25519)
		copyPublicValidatorKey = crypto.PubKey(copyKeyBytes)
	default:
		// TODO: [ben] add error return to all these calls
		copyPublicValidatorKey = nil
	}
	return copyPublicValidatorKey
}

func (tendermint *Tendermint) Events() edb_event.EventEmitter {
	return edb_event.NewEvents(tendermint.tmintNode.EventSwitch(), tendermint.logger)
}

func (tendermint *Tendermint) BroadcastTransaction(transaction []byte,
	callback func(*tmsp_types.Response)) error {
	return tendermint.tmintNode.MempoolReactor().BroadcastTx(transaction, callback)
}

func (tendermint *Tendermint) ListUnconfirmedTxs(
	maxTxs int) ([]txs.Tx, error) {
	tendermintTxs := tendermint.tmintNode.MempoolReactor().Mempool.Reap(maxTxs)
	transactions := make([]txs.Tx, len(tendermintTxs))
	for i, txBytes := range tendermintTxs {
		tx, err := txs.DecodeTx(txBytes)
		if err != nil {
			return nil, err
		}
		transactions[i] = tx
	}
	return transactions, nil
}

func (tendermint *Tendermint) ListValidators() []consensus_types.Validator {
	return consensus_types.FromTendermintValidators(tendermint.tmintNode.
		ConsensusState().Validators.Validators)
}

func (tendermint *Tendermint) ConsensusState() *consensus_types.ConsensusState {
	return consensus_types.FromRoundState(tendermint.tmintNode.ConsensusState().
		GetRoundState())
}

func (tendermint *Tendermint) PeerConsensusStates() map[string]string {
	peers := tendermint.tmintNode.Switch().Peers().List()
	peerConsensusStates := make(map[string]string,
		len(peers))
	for _, peer := range peers {
		peerState := peer.Data.Get(tendermint_types.PeerStateKey).(*tendermint_consensus.PeerState)
		peerRoundState := peerState.GetRoundState()
		// TODO: implement a proper mapping, this is a nasty way of marshalling
		// to JSON
		peerConsensusStates[peer.Key] = string(wire.JSONBytes(peerRoundState))
	}
	return peerConsensusStates
}

//------------------------------------------------------------------------------
// Helper functions

// func marshalConfigToDisk(filePath string, tendermintConfig *viper.Viper) error {
//
//   tendermintConfig.Unmarshal
//   // marshal interface to toml bytes
//   bytesConfig, err := toml.Marshal(tendermintConfig)
//   if err != nil {
//     return fmt.Fatalf("Failed to marshal Tendermint configuration to bytes: %v",
//       err)
//   }
//   return files.WriteAndBackup(filePath, bytesConfig)
// }
