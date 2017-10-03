package tendermint

import (
	"time"

	"fmt"

	acm "github.com/hyperledger/burrow/account"
	bcm "github.com/hyperledger/burrow/blockchain"
	"github.com/hyperledger/burrow/consensus/tendermint/abci"
	"github.com/hyperledger/burrow/consensus/tendermint/query"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging"
	logging_types "github.com/hyperledger/burrow/logging/types"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/rpc/core"
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

func NewNode(conf *config.Config, privValidator tm_types.PrivValidator, abciApp abci_types.Application,
	genDocProvider node.GenesisDocProvider, logger logging_types.InfoTraceLogger) (*node.Node, error) {

	return node.NewNode(conf, privValidator, proxy.NewLocalClientCreator(abciApp), genDocProvider,
		node.DefaultDBProvider, NewLogger(logger))
}

func Init() {
}

func LaunchGenesisValidator(conf *config.Config, logger logging_types.InfoTraceLogger) error {
	// Validator setup
	validatorAccount := acm.GeneratePrivateAccount()
	genesisDoc := GenerateGenesis(conf.ChainID, validatorAccount)
	burrowGenesis := BurrowGenesis(genesisDoc)
	privValidator := NewPrivValidatorMemory(validatorAccount)

	// Connect Burrow genesis state root with Tendermint's
	genesisDoc.AppHash = burrowGenesis.Hash()

	logger = logging.WithScope(logger, "LaunchGenesisValidator")
	stateDB := dbm.NewDB("burrow_state", dbm.GoLevelDBBackendStr, conf.DBDir())
	state := execution.MakeGenesisState(stateDB, burrowGenesis)
	state.Save()

	blockchain := bcm.NewBlockchain(burrowGenesis)
	eventEmitter := event.NewEvents(events.NewEventSwitch(), logger)
	checker := execution.NewBatchChecker(state, genesisDoc.ChainID, blockchain, logger)
	committer := execution.NewBatchCommitter(state, genesisDoc.ChainID, blockchain, eventEmitter, logger)
	validatorNode, err := NewNode(conf, privValidator, abci.NewApp(blockchain, checker, committer, logger),
		func() (*tm_types.GenesisDoc, error) {
			return genesisDoc, nil
		}, logger)

	if err != nil {
		return err
	}

	txCodec := txs.NewGoWireCodec()
	transactor := execution.NewTransactor(blockchain, state, eventEmitter,
		BroadcastTxAsyncFunc(validatorNode, txCodec))

	nameReg := execution.NewNameReg(state, blockchain)

	service := core.NewService(
		state,
		eventEmitter,
		nameReg,
		blockchain,
		transactor,
		query.NewNodeView(validatorNode, txCodec),
		logger,
	)

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

func GenerateGenesis(chainID string, privateAccount acm.PrivateAccount) *tm_types.GenesisDoc {
	return &tm_types.GenesisDoc{
		GenesisTime: time.Now(),
		ChainID:     chainID,
		Validators: []tm_types.GenesisValidator{{
			PubKey: privateAccount.PubKey(),
			Power:  GenesisValidatorPower,
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
