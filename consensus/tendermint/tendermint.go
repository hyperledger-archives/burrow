package tendermint

import (
	"time"

	"github.com/hyperledger/burrow/account"
	bcm "github.com/hyperledger/burrow/blockchain"
	"github.com/hyperledger/burrow/consensus/tendermint/abci"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging"
	logging_types "github.com/hyperledger/burrow/logging/types"
	"github.com/hyperledger/burrow/permission"
	abci_types "github.com/tendermint/abci/types"
	"github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/proxy"
	tm_types "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/events"
)

const GenesisValidatorAmount = 1 << 10

func NewNode(conf *config.Config, privValidator tm_types.PrivValidator, abciApp abci_types.Application,
	logger logging_types.InfoTraceLogger) *node.Node {

	return node.NewNode(conf, privValidator, proxy.NewLocalClientCreator(abciApp), NewLogger(logger),
		node.DefaultDBProvider)
}

func LaunchGenesisValidator(conf *config.Config, logger logging_types.InfoTraceLogger) error {
	privValidator, genesisDoc := GenerateGenesis(conf.ChainID)
	burrowGenesis := BurrowGenesis(genesisDoc)

	genesisDoc.AppHash = burrowGenesis.Hash()

	genesisDoc.SaveAs(conf.GenesisFile())
	privValidator.SetFile(conf.GenesisFile())
	logger = logging.WithScope(logger, "LaunchGenesisValidator")
	stateDB := dbm.NewDB("burrow_state", dbm.GoLevelDBBackendStr, conf.DBDir())
	state := execution.MakeGenesisState(stateDB, burrowGenesis)
	state.Save()

	blockchain := bcm.NewBlockchainFromGenesisDoc(burrowGenesis)
	eventSwitch := events.NewEventSwitch()
	checker := execution.NewBatchChecker(state, blockchain, logger)
	committer := execution.NewBatchCommitter(state, blockchain, eventSwitch, logger)

	validatorNode := NewNode(conf, privValidator,
		abci.NewApp(blockchain, checker, committer, logger), logger)

	_, err := validatorNode.Start()
	if err != nil {
		return err
	}
	validatorNode.NodeInfo()
	validatorNode.RunForever()
	return nil
}

func GenerateGenesis(burrowGenesis *genesis.GenesisDoc) (*tm_types.PrivValidator, *tm_types.GenesisDoc) {
	privValidator := tm_types.GenPrivValidato()
	return privValidator, &tm_types.GenesisDoc{
		GenesisTime: time.Now(),
		ChainID:     burrowGenesis.ChainID,
		Validators: []tm_types.GenesisValidator{{
			PubKey: privValidator.PubKey,
			Amount: GenesisValidatorAmount,
			Name:   "GenesisValidator",
		}},
	}
}

func BurrowGenesis(tmGenesisDoc *tm_types.GenesisDoc) *genesis.GenesisDoc {
	accs := make([]*genesis.GenesisAccount, len(tmGenesisDoc.Validators))
	vals := make([]*genesis.GenesisValidator, len(tmGenesisDoc.Validators))
	for i, tmVal := range tmGenesisDoc.Validators {
		// Copy before taking a pointer
		defaultPermissions := permission.DefaultAccountPermissions
		accs[i] = &genesis.GenesisAccount{
			BasicAccount: genesis.BasicAccount{
				Address: account.MustAddressFromBytes(tmVal.PubKey.Address()),
				Amount:  tmVal.Amount,
			},
			Name:        tmVal.Name,
			Permissions: &defaultPermissions,
		}
		vals[i] = &genesis.GenesisValidator{
			PubKey:   tmVal.PubKey,
			Name:     tmVal.Name,
			Amount:   tmVal.Amount,
			UnbondTo: []genesis.BasicAccount{accs[i].BasicAccount},
		}
	}
	return genesis.MakeGenesisDocFromAccounts(tmGenesisDoc.ChainID, accs, vals)
}
