package tendermint

import (
	"time"

	"fmt"

	acm "github.com/hyperledger/burrow/account"
	bcm "github.com/hyperledger/burrow/blockchain"
	"github.com/hyperledger/burrow/consensus/tendermint/abci"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging"
	logging_types "github.com/hyperledger/burrow/logging/types"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/txs"
	abci_types "github.com/tendermint/abci/types"
	"github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/proxy"
	tm_types "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/events"
)

const GenesisValidatorPower = 1 << 10

type NodeDeps struct {
	*config.Config
	tm_types.PrivValidator
	*tm_types.GenesisDoc
	bcm.MutableBlockchain
	*execution.State
	event.EventEmitter
	logging_types.InfoTraceLogger
}

func deps() {
	NodeDeps{
		InfoTraceLogger:,
		EventEmitter:,
	}
}

func NewNode(
	conf *config.Config,
	privValidator tm_types.PrivValidator,
	genesisDoc *tm_types.GenesisDoc,
	blockchain bcm.MutableBlockchain,
	state *execution.State,
	eventEmitter event.EventEmitter,
	logger logging_types.InfoTraceLogger) (*node.Node, error) {

	checker := execution.NewBatchChecker(state, genesisDoc.ChainID, blockchain, logger)
	committer := execution.NewBatchCommitter(state, genesisDoc.ChainID, blockchain, eventEmitter, logger)
	app := abci.NewApp(blockchain, checker, committer, logger)
	return node.NewNode(conf, privValidator,
		proxy.NewLocalClientCreator(app),
		func() (*tm_types.GenesisDoc, error) {
			return genesisDoc, nil
		},
		node.DefaultDBProvider,
		NewLogger(logger))
}

func Init() {
}

func LaunchGenesisValidator(conf *config.Config, logger logging_types.InfoTraceLogger) error {
	// disable Tendermint's RPC
	conf.RPC.ListenAddress = ""
	// Validator setup
	validatorAccount := acm.GeneratePrivateAccount()
	genesisDoc := GenerateGenesis(conf.ChainID, validatorAccount)
	burrowGenesis := BurrowGenesis(genesisDoc)
	privValidator := NewPrivValidatorMemory(validatorAccount)

	// Connect Burrow genesis state root with Tendermint's
	genesisDoc.AppHash = burrowGenesis.Hash()

	// Logger
	logger = logging.WithScope(logger, "LaunchGenesisValidator")

	// State
	stateDB := dbm.NewDB("burrow_state", dbm.GoLevelDBBackendStr, conf.DBDir())
	state := execution.MakeGenesisState(stateDB, burrowGenesis)
	state.Save()

	blockchain := bcm.NewBlockchain(burrowGenesis)
	eventSwitch := events.NewEventSwitch()
	eventEmitter := event.NewEvents(eventSwitch, logger)

	validatorNode, err := NewNode(conf, privValidator, genesisDoc, blockchain, state, eventEmitter, logger)
	if err != nil {
		return err
	}

	_, err = validatorNode.Start()
	if err != nil {
		return err
	}
	validatorNode.RunForever()
	return nil
}

func BroadcastTxAsyncFunc(validator *node.Node, txEncoder txs.Encoder) func(tx txs.Tx, callback func(res *abci_types.Response)) error {
	return func(tx txs.Tx, callback func(res *abci_types.Response)) error {

		txBytes, err := txEncoder.EncodeTx(tx)
		if err != nil {
			return fmt.Errorf("error encoding transaction: %v", err)
		}

		err = validator.MempoolReactor().BroadcastTx(txBytes, callback)
		if err != nil {
			return fmt.Errorf("error broadcasting transaction: %v", err)
		}
		return nil
	}
}

func DeriveGenesisDoc(burrowGenesisDoc *genesis.GenesisDoc) *tm_types.GenesisDoc {
	accs := make([]*genesis.GenesisAccount, len(tmGenesisDoc.Validators))
	vals := make([]*genesis.GenesisValidator, len(tmGenesisDoc.Validators))
	for i, tmVal := range tmGenesisDoc.Validators {
		// Copy before taking a pointer
		defaultPermissions := permission.DefaultAccountPermissions
		accs[i] = &genesis.GenesisAccount{
			BasicAccount: genesis.BasicAccount{
				Address: acm.MustAddressFromBytes(tmVal.PubKey.Address()),
				Amount:  tmVal.Power,
			},
			Name:        tmVal.Name,
			Permissions: &defaultPermissions,
		}
		vals[i] = &genesis.GenesisValidator{
			PubKey:   tmVal.PubKey,
			Name:     tmVal.Name,
			Amount:   tmVal.Power,
			UnbondTo: []genesis.BasicAccount{accs[i].BasicAccount},
		}
	}
	return genesis.MakeGenesisDocFromAccounts(tmGenesisDoc.ChainID, accs, vals)
}
