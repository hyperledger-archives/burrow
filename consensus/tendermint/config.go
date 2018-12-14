package tendermint

import (
	"net/url"
	"strings"

	"github.com/hyperledger/burrow/consensus/tendermint/abci"
	tm_config "github.com/tendermint/tendermint/config"
)

// Burrow's view on Tendermint's config. Since we operate as a Tendermint harness not all configuration values
// are applicable, we may not allow some values to specified, or we may not allow some to be set independently.
// So this serves as a layer of indirection over Tendermint's real config that we derive from ours.
type BurrowTendermintConfig struct {
	// Initial peers we connect to for peer exchange
	Seeds string
	// Whether this node should crawl the network looking for new peers - disconnecting to peers after it has shared addresses
	SeedMode bool
	// Peers to which we automatically connect
	PersistentPeers string
	ListenAddress   string
	// Optional external that nodes may provide with their NodeInfo
	ExternalAddress string
	// Set true for strict address routability rules
	// Set false for private or local networks
	AddrBookStrict bool
	Moniker        string
	TendermintRoot string
	// Peers ID or address this node is authorize to sync with
	AuthorizedPeers string

	// EmptyBlocks mode and possible interval between empty blocks in seconds
	CreateEmptyBlocks         bool
	CreateEmptyBlocksInterval int
}

func DefaultBurrowTendermintConfig() *BurrowTendermintConfig {
	tmDefaultConfig := tm_config.DefaultConfig()
	return &BurrowTendermintConfig{
		ListenAddress:             tmDefaultConfig.P2P.ListenAddress,
		ExternalAddress:           tmDefaultConfig.P2P.ExternalAddress,
		TendermintRoot:            ".burrow",
		CreateEmptyBlocks:         tmDefaultConfig.Consensus.CreateEmptyBlocks,
		CreateEmptyBlocksInterval: tmDefaultConfig.Consensus.CreateEmptyBlocksInterval,
	}
}

func (btc *BurrowTendermintConfig) TendermintConfig() *tm_config.Config {
	conf := tm_config.DefaultConfig()
	if btc != nil {
		// We may need to expose more of the P2P/Consensus/Mempool options, but I'd like to keep the configuration
		// minimal
		conf.RootDir = btc.TendermintRoot
		conf.Consensus.RootDir = btc.TendermintRoot
		conf.Consensus.CreateEmptyBlocks = btc.CreateEmptyBlocks
		conf.Consensus.CreateEmptyBlocksInterval = btc.CreateEmptyBlocksInterval
		conf.Mempool.RootDir = btc.TendermintRoot
		conf.P2P.RootDir = btc.TendermintRoot
		conf.P2P.Seeds = btc.Seeds
		conf.P2P.SeedMode = btc.SeedMode
		conf.P2P.PersistentPeers = btc.PersistentPeers
		conf.P2P.ListenAddress = btc.ListenAddress
		conf.P2P.ExternalAddress = btc.ExternalAddress
		conf.P2P.AddrBookStrict = btc.AddrBookStrict
		conf.Moniker = btc.Moniker
		// Unfortunately this stops metrics from being used at all
		conf.Instrumentation.Prometheus = false
		conf.FilterPeers = btc.AuthorizedPeers != ""
	}
	// Disable Tendermint RPC
	conf.RPC.ListenAddress = ""
	return conf
}

func (btc *BurrowTendermintConfig) DefaultAuthorizedPeersProvider() abci.PeersFilterProvider {
	var authorizedPeersID, authorizedPeersAddress []string

	authorizedPeersAddrOrID := strings.Split(btc.AuthorizedPeers, ",")
	for _, authorizedPeerAddrOrID := range authorizedPeersAddrOrID {
		_, err := url.Parse(authorizedPeerAddrOrID)
		isNodeAddress := err != nil
		if isNodeAddress {
			authorizedPeersAddress = append(authorizedPeersAddress, authorizedPeerAddrOrID)
		} else {
			authorizedPeersID = append(authorizedPeersID, authorizedPeerAddrOrID)
		}
	}

	return func() ([]string, []string) {
		return authorizedPeersID, authorizedPeersAddress
	}
}
