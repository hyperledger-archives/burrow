package tendermint

import (
	acm "github.com/hyperledger/burrow/account"
	tm_config "github.com/tendermint/tendermint/config"
)

// Burrow's view on Tendermint's config. Since we operate as a Tendermint harness not all configuration values
// are applicable, we may not allow some values to specified, or we may not allow some to be set independently.
// So this serves as a layer of indirection over Tendermint's real config that we derive from ours.
type BurrowTendermintConfig struct {
	Seeds            string      `toml:"seeds"`
	ListenAddress    string      `toml:"listen_address"`
	Moniker          string      `toml:"moniker"`
	ValidatorAddress acm.Address `toml:"validator_address"`
}

func (btc *BurrowTendermintConfig) TendermintConfig() *tm_config.Config {
	conf := tm_config.DefaultConfig()
	conf.P2P.Seeds = btc.Seeds
	conf.P2P.ListenAddress = btc.ListenAddress
	conf.Moniker = btc.Moniker
	// Disable Tendermint RPC
	conf.RPC.ListenAddress = ""
	return conf
}
