package tendermint

import (
	"os"

	tm_config "github.com/tendermint/tendermint/config"
)

// Burrow's view on Tendermint's config. Since we operate as a Tendermint harness not all configuration values
// are applicable, we may not allow some values to specified, or we may not allow some to be set independently.
// So this serves as a layer of indirection over Tendermint's real config that we derive from ours.
type BurrowTendermintConfig struct {
	// Initial peers we connect to for peer exchange
	Seeds string
	// Peers to which we automatically connect
	PersistentPeers string
	ListenAddress   string
	Moniker         string
	TendermintRoot  string
}

func DefaultBurrowTendermintConfig() *BurrowTendermintConfig {
	tmDefaultConfig := tm_config.DefaultConfig()
	return &BurrowTendermintConfig{
		ListenAddress:  tmDefaultConfig.P2P.ListenAddress,
		TendermintRoot: ".burrow",
	}
}

func (btc *BurrowTendermintConfig) TendermintConfig() *tm_config.Config {
	conf := tm_config.DefaultConfig()
	if btc != nil {
		// We may need to expose more of the P2P/Consensus/Mempool options, but I'd like to keep the configuration
		// minimal
		os.MkdirAll(btc.TendermintRoot+"/config", 0755) /// Create directory for tendermint config files
		conf.SetRoot(btc.TendermintRoot)                /// set tendermint root file (--home)

		conf.P2P.Seeds = btc.Seeds
		conf.P2P.PersistentPeers = btc.PersistentPeers
		conf.P2P.ListenAddress = btc.ListenAddress
		conf.Moniker = btc.Moniker
	}
	conf.RPC.ListenAddress = "tcp://localhost:0"
	return conf
}
