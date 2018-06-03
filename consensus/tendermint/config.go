package tendermint

import (
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
		conf.RootDir = btc.TendermintRoot
		conf.Consensus.RootDir = btc.TendermintRoot
		conf.Mempool.RootDir = btc.TendermintRoot
		conf.P2P.RootDir = btc.TendermintRoot
		conf.P2P.Seeds = btc.Seeds
		conf.P2P.PersistentPeers = btc.PersistentPeers
		conf.P2P.ListenAddress = btc.ListenAddress
		conf.Moniker = btc.Moniker
	}
	// Disable Tendermint RPC
	conf.RPC.ListenAddress = ""
	return conf
}
