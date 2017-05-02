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

package tendermint

import (
	"fmt"
	"path"
	"strings"

	tendermint_version "github.com/hyperledger/burrow/consensus/tendermint/version"
	abci_types "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
	p2p "github.com/tendermint/go-p2p"
	tendermint_consensus "github.com/tendermint/tendermint/consensus"
	node "github.com/tendermint/tendermint/node"
	proxy "github.com/tendermint/tendermint/proxy"
	tendermint_types "github.com/tendermint/tendermint/types"

	edb_event "github.com/hyperledger/burrow/event"

	config "github.com/hyperledger/burrow/config"
	manager_types "github.com/hyperledger/burrow/manager/types"
	// files  "github.com/hyperledger/burrow/files"
	"errors"

	blockchain_types "github.com/hyperledger/burrow/blockchain/types"
	consensus_types "github.com/hyperledger/burrow/consensus/types"
	"github.com/hyperledger/burrow/logging"
	logging_types "github.com/hyperledger/burrow/logging/types"
	"github.com/hyperledger/burrow/txs"
	"github.com/tendermint/go-wire"
)

type Tendermint struct {
	tmintNode   *node.Node
	tmintConfig *TendermintConfig
	chainId     string
	logger      logging_types.InfoTraceLogger
}

// Compiler checks to ensure Tendermint successfully implements
// burrow/definitions Consensus and Blockchain
var _ consensus_types.ConsensusEngine = (*Tendermint)(nil)
var _ blockchain_types.Blockchain = (*Tendermint)(nil)

func NewTendermint(moduleConfig *config.ModuleConfig,
	application manager_types.Application,
	logger logging_types.InfoTraceLogger) (*Tendermint, error) {
	// re-assert proper configuration for module
	if moduleConfig.Version != tendermint_version.GetTendermintVersion().GetMinorVersionString() {
		return nil, fmt.Errorf("Version string %s did not match %s",
			moduleConfig.Version, tendermint_version.GetTendermintVersion().GetMinorVersionString())
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
	// TODO: [ben] implement the signer for Private validator over monax-keys
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

	newNode := node.NewNode(tmintConfig, privateValidator,
		proxy.NewLocalClientCreator(application))

	// TODO: [ben] delay starting the node to a different function, to hand
	// control over events to Core
	if started, err := newNode.Start(); !started {
		newNode.Stop()
		if err != nil {
			return nil, fmt.Errorf("Failed to start Tendermint consensus node: %v", err)
		}
		return nil, errors.New("Failed to start Tendermint consensus node, " +
			"probably because it is already started, see logs")

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
	callback func(*abci_types.Response)) error {
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

// Allow for graceful shutdown of node. Returns whether the node was stopped.
func (tendermint *Tendermint) Stop() bool {
	return tendermint.tmintNode.Stop()
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
