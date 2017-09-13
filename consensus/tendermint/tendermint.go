package tendermint

import (
	"time"

	"crypto/rand"

	"github.com/hyperledger/burrow/consensus/tendermint/abci"
	logging_types "github.com/hyperledger/burrow/logging/types"
	"github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/proxy"
	"github.com/tendermint/tendermint/types"
)

const GenesisValidatorAmount = 1 << 10
const AppHashLength = 32

func LaunchGenesisValidator(conf *config.Config, logger logging_types.InfoTraceLogger) error {
	appHash := make([]byte, AppHashLength)
	_, err := rand.Read(appHash)
	if err != nil {
		return err
	}
	privValidator, genesisDoc := Generate(conf.ChainID, appHash)
	genesisDoc.SaveAs(conf.GenesisFile())
	privValidator.SetFile(conf.GenesisFile())
	tmLogger := NewLogger(logger)
	validatorNode := node.NewNode(conf, privValidator,
		proxy.NewLocalClientCreator(abci.NewApp(genesisDoc.AppHash)), tmLogger)
	_, err = validatorNode.Start()
	if err != nil {
		return err
	}
	validatorNode.RunForever()
	return nil
}

func Generate(chainID string, appHash []byte) (*types.PrivValidator, *types.GenesisDoc) {
	privValidator := types.GenPrivValidator()
	return privValidator, &types.GenesisDoc{
		GenesisTime: time.Now(),
		ChainID:     chainID,
		Validators: []types.GenesisValidator{{
			PubKey: privValidator.PubKey,
			Amount: GenesisValidatorAmount,
			Name:   "GenesisValidator",
		}},
		AppHash: appHash,
	}
}
