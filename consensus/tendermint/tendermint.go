package tendermint

import (
	"fmt"

	bcm "github.com/hyperledger/burrow/blockchain"
	"github.com/hyperledger/burrow/consensus/tendermint/abci"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging/structure"
	logging_types "github.com/hyperledger/burrow/logging/types"
	"github.com/hyperledger/burrow/txs"
	abci_types "github.com/tendermint/abci/types"
	"github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/proxy"
	tm_types "github.com/tendermint/tendermint/types"
)

func NewNode(
	conf *config.Config,
	privValidator tm_types.PrivValidator,
	genesisDoc *tm_types.GenesisDoc,
	blockchain bcm.MutableBlockchain,
	checker execution.BatchExecutor,
	committer execution.BatchCommitter,
	logger logging_types.InfoTraceLogger) (*node.Node, error) {

	// disable Tendermint's RPC
	conf.RPC.ListenAddress = ""

	app := abci.NewApp(blockchain, checker, committer, logger)
	return node.NewNode(conf, privValidator,
		proxy.NewLocalClientCreator(app),
		func() (*tm_types.GenesisDoc, error) {
			return genesisDoc, nil
		},
		node.DefaultDBProvider,
		NewLogger(logger.WithPrefix(structure.ComponentKey, "Tendermint").
			With(structure.ScopeKey, "tendermint.NewNode")))
}

func BroadcastTxAsyncFunc(validator *node.Node, txEncoder txs.Encoder) func(tx txs.Tx,
	callback func(res *abci_types.Response)) error {

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
	validators := make([]tm_types.GenesisValidator, len(burrowGenesisDoc.Validators))
	for i, validator := range burrowGenesisDoc.Validators {
		validators[i] = tm_types.GenesisValidator{
			PubKey: validator.PubKey,
			Name:   validator.Name,
			Power:  int64(validator.Amount),
		}
	}
	return &tm_types.GenesisDoc{
		ChainID:     burrowGenesisDoc.ChainID(),
		GenesisTime: burrowGenesisDoc.GenesisTime,
		Validators:  validators,
		AppHash:     burrowGenesisDoc.Hash(),
	}
}
