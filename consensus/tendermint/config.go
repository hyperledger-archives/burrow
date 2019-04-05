package tendermint

import (
	"math"
	"net/url"
	"strings"
	"time"

	"github.com/hyperledger/burrow/consensus/abci"
	tmConfig "github.com/tendermint/tendermint/config"
)

// Burrow's view on Tendermint's config. Since we operate as a Tendermint harness not all configuration values
// are applicable, we may not allow some values to specified, or we may not allow some to be set independently.
// So this serves as a layer of indirection over Tendermint's real config that we derive from ours.
type BurrowTendermintConfig struct {
	Enabled bool
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
	// Peers ID or address this node is authorize to sync with
	AuthorizedPeers string
	// EmptyBlocks mode and possible interval between empty blocks in seconds
	CreateEmptyBlocks         bool
	CreateEmptyBlocksInterval time.Duration
}

func DefaultBurrowTendermintConfig() *BurrowTendermintConfig {
	tmDefaultConfig := tmConfig.DefaultConfig()
	return &BurrowTendermintConfig{
		Enabled:                   true,
		ListenAddress:             tmDefaultConfig.P2P.ListenAddress,
		ExternalAddress:           tmDefaultConfig.P2P.ExternalAddress,
		CreateEmptyBlocks:         tmDefaultConfig.Consensus.CreateEmptyBlocks,
		CreateEmptyBlocksInterval: tmDefaultConfig.Consensus.CreateEmptyBlocksInterval,
	}
}

func (btc *BurrowTendermintConfig) Config(rootDir string, timeoutFactor float64) *tmConfig.Config {
	conf := tmConfig.DefaultConfig()
	// We expose Tendermint config as required, but try to give fewer levers to pull where possible
	if btc != nil {
		conf.RootDir = rootDir
		conf.Mempool.RootDir = rootDir
		conf.Consensus.RootDir = rootDir

		// Consensus
		conf.Consensus.CreateEmptyBlocks = btc.CreateEmptyBlocks
		conf.Consensus.CreateEmptyBlocksInterval = btc.CreateEmptyBlocksInterval
		// Assume Tendermint has some mutually consistent values, assume scaling them linearly makes sense
		conf.Consensus.TimeoutPropose = scaleTimeout(timeoutFactor, conf.Consensus.TimeoutPropose)
		conf.Consensus.TimeoutProposeDelta = scaleTimeout(timeoutFactor, conf.Consensus.TimeoutProposeDelta)
		conf.Consensus.TimeoutPrevote = scaleTimeout(timeoutFactor, conf.Consensus.TimeoutPrevote)
		conf.Consensus.TimeoutPrevoteDelta = scaleTimeout(timeoutFactor, conf.Consensus.TimeoutPrevoteDelta)
		conf.Consensus.TimeoutPrecommit = scaleTimeout(timeoutFactor, conf.Consensus.TimeoutPrecommit)
		conf.Consensus.TimeoutPrecommitDelta = scaleTimeout(timeoutFactor, conf.Consensus.TimeoutPrecommitDelta)
		conf.Consensus.TimeoutCommit = scaleTimeout(timeoutFactor, conf.Consensus.TimeoutCommit)

		// P2P
		conf.Moniker = btc.Moniker
		conf.P2P.RootDir = rootDir
		conf.P2P.Seeds = btc.Seeds
		conf.P2P.SeedMode = btc.SeedMode
		conf.P2P.PersistentPeers = btc.PersistentPeers
		conf.P2P.ListenAddress = btc.ListenAddress
		conf.P2P.ExternalAddress = btc.ExternalAddress
		conf.P2P.AddrBookStrict = btc.AddrBookStrict
		// We use this in tests and I am not aware of a strong reason to reject nodes on the same IP with different ports
		conf.P2P.AllowDuplicateIP = true

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

func scaleTimeout(factor float64, timeout time.Duration) time.Duration {
	if factor == 0 {
		return timeout
	}
	return time.Duration(math.Round(factor * float64(timeout)))
}
